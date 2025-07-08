local M = {}
local config = require("debugstory.config")

-- Plugin state
local session_active = false
local session_id = nil
local go_process = nil

-- Helper function to execute Go binary
local function exec_go_command(cmd, args)
	local go_binary = vim.fn.stdpath("data") .. "/lazy/debugstory.nvim/bin/debugstory"
	local full_cmd = go_binary .. " " .. cmd

	if args then
		for _, arg in ipairs(args) do
			full_cmd = full_cmd .. " " .. vim.fn.shellescape(arg)
		end
	end

	local result = vim.fn.system(full_cmd)
	return result
end

-- Start a new debug session
function M.start_session(project_name)
	if session_active then
		vim.notify("Debug session already active", vim.log.levels.WARN)
		return
	end

	project_name = project_name or vim.fn.fnamemodify(vim.fn.getcwd(), ":t")
	session_id = os.time() .. "_" .. project_name

	local result = exec_go_command("start", {
		session_id,
		vim.fn.getcwd(),
		config.get().save_path,
		config.get().output_format,
	})

	if vim.v.shell_error == 0 then
		session_active = true
		vim.notify("Debug session started: " .. session_id, vim.log.levels.INFO)
		M.setup_autocommands()
	else
		vim.notify("Failed to start debug session: " .. result, vim.log.levels.ERROR)
	end
end

-- End current debug session
function M.end_session()
	if not session_active then
		vim.notify("No active debug session", vim.log.levels.WARN)
		return
	end

	local result = exec_go_command("end", { session_id, config.get().save_path })

	if vim.v.shell_error == 0 then
		session_active = false
		session_id = nil
		vim.notify("Debug session ended and saved", vim.log.levels.INFO)
		M.cleanup_autocommands()
	else
		vim.notify("Failed to end debug session: " .. result, vim.log.levels.ERROR)
	end
end

-- Add annotation to current session
function M.add_annotation(note)
	if not session_active then
		vim.notify("No active debug session", vim.log.levels.WARN)
		return
	end

	if not note then
		note = vim.fn.input("Annotation: ")
	end

	if note and note ~= "" then
		local result = exec_go_command("annotate", { session_id, config.get().save_path, note })
		if vim.v.shell_error == 0 then
			vim.notify("Annotation added", vim.log.levels.INFO)
		else
			vim.notify("Failed to add annotation: " .. result, vim.log.levels.ERROR)
		end
	end
end

-- Record file edit
function M.record_edit(bufnr, changedtick)
	if not session_active then
		return
	end

	local filename = vim.api.nvim_buf_get_name(bufnr)
	if filename == "" then
		return
	end

	local cursor_pos = vim.api.nvim_win_get_cursor(0)
	local line_count = vim.api.nvim_buf_line_count(bufnr)

	exec_go_command("record-edit", {
		session_id,
		config.get().save_path,
		filename,
		tostring(cursor_pos[1]),
		tostring(cursor_pos[2]),
		tostring(line_count),
		tostring(changedtick),
	})
end

-- Record terminal command
function M.record_terminal_command(cmd)
	if not session_active then
		return
	end

	exec_go_command("record-terminal", { session_id, config.get().save_path, cmd })
end

-- Setup autocommands for recording
function M.setup_autocommands()
	local group = vim.api.nvim_create_augroup("DebugStory", { clear = true })

	-- Record file changes
	vim.api.nvim_create_autocmd({ "TextChanged", "TextChangedI" }, {
		group = group,
		callback = function()
			M.record_edit(vim.api.nvim_get_current_buf(), vim.api.nvim_buf_get_changedtick(0))
		end,
	})

	-- Record cursor movements
	vim.api.nvim_create_autocmd("CursorMoved", {
		group = group,
		callback = function()
			if session_active then
				local cursor_pos = vim.api.nvim_win_get_cursor(0)
				local filename = vim.api.nvim_buf_get_name(0)
				exec_go_command("record-cursor", {
					session_id,
					config.get().save_path,
					filename,
					tostring(cursor_pos[1]),
					tostring(cursor_pos[2]),
				})
			end
		end,
	})

	-- Auto-save on exit
	if config.get().auto_save_on_exit then
		vim.api.nvim_create_autocmd("VimLeavePre", {
			group = group,
			callback = function()
				if session_active then
					M.end_session()
				end
			end,
		})
	end
end

-- Clean up autocommands
function M.cleanup_autocommands()
	vim.api.nvim_clear_autocmds({ group = "DebugStory" })
end

-- Get session status
function M.get_status()
	if session_active then
		return {
			active = true,
			session_id = session_id,
			save_path = config.get().save_path,
		}
	else
		return { active = false }
	end
end

-- List available sessions
function M.list_sessions()
	local result = exec_go_command("list", { config.get().save_path })
	if vim.v.shell_error == 0 then
		return vim.split(result, "\n")
	else
		vim.notify("Failed to list sessions: " .. result, vim.log.levels.ERROR)
		return {}
	end
end

-- Resume a session
function M.resume_session(session_name)
	if session_active then
		vim.notify("Please end current session first", vim.log.levels.WARN)
		return
	end

	local result = exec_go_command("resume", { session_name, config.get().save_path })
	if vim.v.shell_error == 0 then
		session_active = true
		session_id = session_name
		vim.notify("Session resumed: " .. session_name, vim.log.levels.INFO)
		M.setup_autocommands()
	else
		vim.notify("Failed to resume session: " .. result, vim.log.levels.ERROR)
	end
end

-- Setup function
function M.setup(opts)
	config.setup(opts)

	-- Create user commands
	vim.api.nvim_create_user_command("DebugStoryStart", function(args)
		M.start_session(args.args ~= "" and args.args or nil)
	end, { nargs = "?", desc = "Start debug session" })

	vim.api.nvim_create_user_command("DebugStoryEnd", function()
		M.end_session()
	end, { desc = "End debug session" })

	vim.api.nvim_create_user_command("DebugStoryAnnotate", function(args)
		M.add_annotation(args.args ~= "" and args.args or nil)
	end, { nargs = "?", desc = "Add annotation to session" })

	vim.api.nvim_create_user_command("DebugStoryStatus", function()
		local status = M.get_status()
		if status.active then
			vim.notify("Active session: " .. status.session_id, vim.log.levels.INFO)
		else
			vim.notify("No active session", vim.log.levels.INFO)
		end
	end, { desc = "Show session status" })

	vim.api.nvim_create_user_command("DebugStoryList", function()
		local sessions = M.list_sessions()
		if #sessions > 0 then
			vim.notify("Available sessions:\n" .. table.concat(sessions, "\n"), vim.log.levels.INFO)
		else
			vim.notify("No sessions found", vim.log.levels.INFO)
		end
	end, { desc = "List available sessions" })

	vim.api.nvim_create_user_command("DebugStoryResume", function(args)
		if args.args == "" then
			vim.notify("Please specify session name", vim.log.levels.WARN)
			return
		end
		M.resume_session(args.args)
	end, { nargs = 1, desc = "Resume debug session" })
end

return M
