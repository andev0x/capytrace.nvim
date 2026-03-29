local M = {}
local config = require("capytrace.config")

-- Plugin state
local session_active = false
local session_id = nil
local go_process = nil
local daemon_chan_id = nil
local request_seq = 0

local function get_go_binary_path()
	local cfg = config.get()
	if cfg.binary_path and cfg.binary_path ~= "" then
		return vim.fn.expand(cfg.binary_path)
	end
	return vim.fn.stdpath("data") .. "/lazy/capytrace.nvim/bin/capytrace"
end

local function binary_exists(path)
	return vim.fn.filereadable(path) == 1
end

local function detect_archive_ext()
	if vim.fn.has("win32") == 1 then
		return "zip"
	end
	return "tar.gz"
end

local function detect_os()
	if vim.fn.has("mac") == 1 then
		return "darwin"
	end
	if vim.fn.has("win32") == 1 then
		return "windows"
	end
	return "linux"
end

local function detect_arch()
	local uname_info = vim.loop.os_uname()
	local uname = (uname_info and uname_info.machine or ""):gsub("%s+", "")
	if uname == "x86_64" then
		return "amd64"
	end
	if uname == "aarch64" then
		return "arm64"
	end
	if uname == "arm64" then
		return "arm64"
	end
	return uname
end

local function extract_archive(archive_path, out_dir)
	if archive_path:sub(-4) == ".zip" then
		return vim.fn.system({ "unzip", "-o", archive_path, "-d", out_dir })
	end
	return vim.fn.system({ "tar", "-xzf", archive_path, "-C", out_dir })
end

local function download_binary()
	local cfg = config.get()
	local go_binary = get_go_binary_path()

	if binary_exists(go_binary) then
		return true
	end

	if not cfg.auto_download_binary then
		return false
	end

	vim.fn.mkdir(vim.fn.fnamemodify(go_binary, ":h"), "p")

	local os_name = detect_os()
	local arch = detect_arch()
	local ext = detect_archive_ext()
	local repo = cfg.github_repo or "andev0x/capytrace.nvim"
	local release_api = "https://api.github.com/repos/" .. repo .. "/releases/latest"
	local release_raw = vim.fn.system({ "curl", "-fsSL", release_api })

	if vim.v.shell_error ~= 0 then
		vim.notify("capytrace: failed to query latest release", vim.log.levels.ERROR)
		return false
	end

	local ok, release = pcall(vim.json.decode, release_raw)
	if not ok or type(release) ~= "table" then
		vim.notify("capytrace: invalid release metadata", vim.log.levels.ERROR)
		return false
	end

	local tag = release.tag_name
	if not tag or tag == "" then
		vim.notify("capytrace: missing release tag", vim.log.levels.ERROR)
		return false
	end

	local version_no_v = tag:gsub("^v", "")
	local artifacts = {
		"capytrace.nvim_" .. tag .. "_" .. os_name .. "_" .. arch .. "." .. ext,
		"capytrace.nvim_" .. version_no_v .. "_" .. os_name .. "_" .. arch .. "." .. ext,
	}
	local tmp_dir = vim.fn.stdpath("cache") .. "/capytrace-download"

	vim.fn.mkdir(tmp_dir, "p")

	local archive_path = nil
	for _, artifact in ipairs(artifacts) do
		local url = "https://github.com/" .. repo .. "/releases/download/" .. tag .. "/" .. artifact
		local candidate = tmp_dir .. "/" .. artifact
		vim.fn.system({ "curl", "-fL", "-o", candidate, url })
		if vim.v.shell_error == 0 then
			archive_path = candidate
			break
		end
	end

	if not archive_path then
		vim.notify("capytrace: failed to download release archive", vim.log.levels.ERROR)
		return false
	end

	extract_archive(archive_path, tmp_dir)
	if vim.v.shell_error ~= 0 then
		vim.notify("capytrace: failed to extract archive", vim.log.levels.ERROR)
		return false
	end

	local source_binary = tmp_dir .. "/capytrace"
	if vim.fn.has("win32") == 1 then
		source_binary = tmp_dir .. "/capytrace.exe"
	end

	if vim.fn.filereadable(source_binary) == 0 then
		vim.notify("capytrace: extracted binary not found", vim.log.levels.ERROR)
		return false
	end

	if vim.loop and vim.loop.fs_copyfile then
		local ok_copy, copy_err = pcall(vim.loop.fs_copyfile, source_binary, go_binary)
		if not ok_copy then
			vim.notify("capytrace: failed to install binary: " .. tostring(copy_err), vim.log.levels.ERROR)
			return false
		end
	else
		vim.fn.system({ "cp", source_binary, go_binary })
		if vim.v.shell_error ~= 0 then
			vim.notify("capytrace: failed to install binary", vim.log.levels.ERROR)
			return false
		end
	end

	if vim.fn.has("win32") == 0 then
		vim.fn.system({ "chmod", "+x", go_binary })
	end

	vim.notify("capytrace: binary downloaded: " .. go_binary, vim.log.levels.INFO)
	return true
end

local function ensure_go_binary()
	local go_binary = get_go_binary_path()
	if binary_exists(go_binary) then
		return go_binary
	end

	if download_binary() then
		if binary_exists(go_binary) then
			return go_binary
		end
	end

	vim.notify("capytrace binary not found: " .. go_binary, vim.log.levels.ERROR)
	return nil
end

local function send_daemon_request(command, args)
	if not go_process or not daemon_chan_id then
		return nil
	end

	request_seq = request_seq + 1
	local req = {
		id = request_seq,
		command = command,
		args = args or {},
	}

	local line = vim.json.encode(req)
	vim.api.nvim_chan_send(daemon_chan_id, line .. "\n")
	return true
end

local function start_daemon()
	if go_process then
		return true
	end

	local go_binary = ensure_go_binary()
	if not go_binary then
		return false
	end

	local stdout_chunks = {}
	local stderr_chunks = {}

	local chan = vim.fn.jobstart({ go_binary, "daemon" }, {
		stdout_buffered = false,
		stderr_buffered = false,
		on_stdout = function(_, data)
			for _, line in ipairs(data) do
				if line ~= "" then
					table.insert(stdout_chunks, line)
				end
			end
		end,
		on_stderr = function(_, data)
			for _, line in ipairs(data) do
				if line ~= "" then
					table.insert(stderr_chunks, line)
				end
			end
		end,
	})

	if chan <= 0 then
		vim.notify("capytrace: failed to start daemon", vim.log.levels.ERROR)
		return false
	end

	go_process = {
		stdout = stdout_chunks,
		stderr = stderr_chunks,
	}
	daemon_chan_id = chan
	return true
end

local function stop_daemon()
	if daemon_chan_id then
		vim.fn.jobstop(daemon_chan_id)
	end
	daemon_chan_id = nil
	go_process = nil
end

-- Helper function to execute Go binary
local function exec_go_command(cmd, args)
	local go_binary = ensure_go_binary()
	if not go_binary then
		return ""
	end
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
		start_daemon()
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
	local report_path = config.get().save_path .. "/" .. session_id .. ".md"

	if vim.v.shell_error == 0 then
		session_active = false
		session_id = nil
		vim.notify("Debug session ended and saved", vim.log.levels.INFO)
		M.cleanup_autocommands()
		if config.get().open_report_on_end and vim.fn.filereadable(report_path) == 1 then
			vim.cmd("edit " .. vim.fn.fnameescape(report_path))
		end
		stop_daemon()
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
		if daemon_chan_id then
			send_daemon_request("annotate", { session_id, config.get().save_path, note })
			vim.notify("Annotation added", vim.log.levels.INFO)
			return
		end

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
	local line_text = vim.api.nvim_buf_get_lines(bufnr, cursor_pos[1] - 1, cursor_pos[1], false)[1] or ""

	if daemon_chan_id then
		send_daemon_request("record-edit", {
			session_id,
			config.get().save_path,
			filename,
			tostring(cursor_pos[1]),
			tostring(cursor_pos[2]),
			tostring(line_count),
			tostring(changedtick),
			line_text,
		})
		return
	end

	exec_go_command("record-edit", {
		session_id,
		config.get().save_path,
		filename,
		tostring(cursor_pos[1]),
		tostring(cursor_pos[2]),
		tostring(line_count),
		tostring(changedtick),
		line_text,
	})
end

-- Record terminal command
function M.record_terminal_command(cmd)
	if not session_active then
		return
	end

	if daemon_chan_id then
		send_daemon_request("record-terminal", { session_id, config.get().save_path, cmd })
		return
	end

	exec_go_command("record-terminal", { session_id, config.get().save_path, cmd })
end

-- Record file open
function M.record_file_open(bufnr)
	if not session_active then
		return
	end

	local filename = vim.api.nvim_buf_get_name(bufnr)
	if filename == "" then
		return
	end

	local filetype = vim.bo[bufnr].filetype
	if daemon_chan_id then
		send_daemon_request("record-file-open", { session_id, config.get().save_path, filename, filetype })
		return
	end

	exec_go_command("record-file-open", { session_id, config.get().save_path, filename, filetype })
end

-- Record LSP diagnostic
function M.record_lsp_diagnostic()
	if not session_active then
		return
	end

	local diagnostics = vim.lsp.diagnostic.get_line_diagnostics()
	if #diagnostics == 0 then
		return
	end

	local filename = vim.api.nvim_buf_get_name(0)
	local cursor_pos = vim.api.nvim_win_get_cursor(0)

	for _, diagnostic in ipairs(diagnostics) do
		if daemon_chan_id then
			send_daemon_request("record-lsp-diagnostic", {
				session_id,
				config.get().save_path,
				filename,
				tostring(cursor_pos[1]),
				tostring(cursor_pos[2]),
				diagnostic.message,
				vim.lsp.protocol.DiagnosticSeverity[diagnostic.severity],
			})
		else
		exec_go_command("record-lsp-diagnostic", {
			session_id,
			config.get().save_path,
			filename,
			tostring(cursor_pos[1]),
			tostring(cursor_pos[2]),
			diagnostic.message,
			vim.lsp.protocol.DiagnosticSeverity[diagnostic.severity],
		})
		end
	end
end

-- Setup autocommands for recording
function M.setup_autocommands()
	local group = vim.api.nvim_create_augroup("capytrace", { clear = true })

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
				if daemon_chan_id then
					send_daemon_request("record-cursor", {
						session_id,
						config.get().save_path,
						filename,
						tostring(cursor_pos[1]),
						tostring(cursor_pos[2]),
					})
				else
					exec_go_command("record-cursor", {
						session_id,
						config.get().save_path,
						filename,
						tostring(cursor_pos[1]),
						tostring(cursor_pos[2]),
					})
				end
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

	-- Log terminal commands
	if config.get().log_events.terminal_commands then
		vim.api.nvim_create_autocmd("TermOpen", {
			group = group,
			callback = function()
				M.record_terminal_command("Terminal opened")
			end,
		})
	end

	-- Log file open
	if config.get().log_events.file_open then
		vim.api.nvim_create_autocmd("BufEnter", {
			group = group,
			callback = function()
				M.record_file_open(vim.api.nvim_get_current_buf())
			end,
		})
	end

	-- Log LSP diagnostics
	if config.get().log_events.lsp_diagnostics then
		vim.api.nvim_create_autocmd("DiagnosticChanged", {
			group = group,
			callback = function()
				M.record_lsp_diagnostic()
			end,
		})
	end
end

-- Clean up autocommands
function M.cleanup_autocommands()
	vim.api.nvim_clear_autocmds({ group = "capytrace" })
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
		start_daemon()
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
	ensure_go_binary()

	-- Create user commands
	vim.api.nvim_create_user_command("CapyTraceStart", function(args)
		M.start_session(args.args ~= "" and args.args or nil)
	end, { nargs = "?", desc = "Start a new debug session" })

	vim.api.nvim_create_user_command("CapyTraceEnd", function()
		M.end_session()
	end, { desc = "End current session" })

	vim.api.nvim_create_user_command("CapyTraceAnnotate", function(args)
		M.add_annotation(args.args ~= "" and args.args or nil)
	end, { nargs = "?", desc = "Add annotation to current session" })

	vim.api.nvim_create_user_command("CapyTraceStatus", function()
		local status = M.get_status()
		if status.active then
			vim.notify("Active session: " .. status.session_id, vim.log.levels.INFO)
		else
			vim.notify("No active session", vim.log.levels.INFO)
		end
	end, { desc = "Show current session status" })

	vim.api.nvim_create_user_command("CapyTraceList", function()
		local sessions = M.list_sessions()
		if #sessions > 0 then
			vim.notify("Available sessions:\n" .. table.concat(sessions, "\n"), vim.log.levels.INFO)
		else
			vim.notify("No sessions found", vim.log.levels.INFO)
		end
	end, { desc = "List all available sessions" })

	vim.api.nvim_create_user_command("CapyTraceResume", function(args)
		if args.args == "" then
			vim.notify("Please specify session name", vim.log.levels.WARN)
			return
		end
		M.resume_session(args.args)
	end, { nargs = 1, desc = "Resume a previous session" })

	vim.api.nvim_create_user_command("CapyTraceSessions", function()
		local ok, telescope = pcall(require, "telescope.builtin")
		if not ok then
			vim.notify("Telescope not found. Install nvim-telescope/telescope.nvim", vim.log.levels.WARN)
			return
		end
		telescope.live_grep({
			search_dirs = { config.get().save_path },
			prompt_title = "CapyTrace Sessions",
			glob_pattern = "*.md",
		})
	end, { desc = "Search CapyTrace session reports" })

	vim.api.nvim_create_user_command("CapyTraceSessionSearch", function()
		local ok = pcall(function()
			require("telescope").extensions.capytrace.sessions()
		end)
		if not ok then
			vim.notify("Load Telescope extension with: require('telescope').load_extension('capytrace')", vim.log.levels.WARN)
		end
	end, { desc = "Telescope search sessions by filename/project" })

	vim.api.nvim_create_user_command("CapyTraceNoteSearch", function()
		local ok = pcall(function()
			require("telescope").extensions.capytrace.notes()
		end)
		if not ok then
			vim.notify("Load Telescope extension with: require('telescope').load_extension('capytrace')", vim.log.levels.WARN)
		end
	end, { desc = "Telescope search session annotations" })
end

return M
