local M = {}

-- Default configuration
local default_config = {
	output_format = "markdown", -- or "json"
	save_path = vim.fn.expand("~/capytrace_logs/"),
	record_terminal = true,
	record_git_diff = true,
	auto_save_on_exit = true,
	max_cursor_events = 100, -- Limit cursor movement recordings
	debounce_ms = 500, -- Debounce time for events
	log_events = {
		terminal_commands = true,
		file_open = true,
		lsp_diagnostics = true,
	},
}

local config = {}

function M.setup(opts)
	config = vim.tbl_deep_extend("force", default_config, opts or {})

	-- Ensure save_path exists
	local save_path = vim.fn.expand(config.save_path)
	if vim.fn.isdirectory(save_path) == 0 then
		vim.fn.mkdir(save_path, "p")
	end

	config.save_path = save_path
end

function M.get()
	return config
end

return M
