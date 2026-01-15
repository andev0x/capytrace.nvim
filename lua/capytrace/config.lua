local M = {}

-- Default configuration
local default_config = {
	output_format = "markdown", -- or "json" or "sqlite"
	save_path = vim.fn.expand("~/capytrace_logs/"),
	record_terminal = true,
	record_git_diff = true,
	auto_save_on_exit = true,
	max_cursor_events = 100, -- Limit cursor movement recordings

	-- Smart Filter configuration (Anti-Spam Cursor Filter)
	filter_threshold = 500, -- Idle threshold in milliseconds (default: 500ms)
	debounce_interval = 200, -- Debounce interval for cursor movements (default: 200ms)

	-- Smart Aggregation configuration (Activity Block Builder)
	aggregation = {
		merge_window = 2000, -- Time window for merging file_edit events in milliseconds (default: 2s)
		idle_threshold = 300000, -- Time to record idle gap in milliseconds (default: 5min)
		flow_velocity_threshold = 10.0, -- Minimum velocity to consider flow state (ticks/sec)
		distraction_files = { -- File patterns that count as distractions
			"NvimTree",
			"copilot-chat",
			"neo-tree",
			"CHADTree",
			"fern",
			"undotree",
		},
		periodic_update_interval = 300000, -- Update SESSION_SUMMARY.md every N milliseconds (default: 5min)
	},

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
