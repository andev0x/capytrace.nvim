local has_telescope, telescope = pcall(require, "telescope")

if not has_telescope then
	return
end

local builtin = require("telescope.builtin")

local extension = {}

extension.sessions = function(opts)
	opts = opts or {}
	local capytrace = require("capytrace")
	local save_path = capytrace.get_status().save_path or require("capytrace.config").get().save_path

	builtin.find_files(vim.tbl_extend("force", {
		prompt_title = "CapyTrace Sessions",
		cwd = save_path,
		find_command = { "rg", "--files", "-g", "*.md" },
	}, opts))
end

extension.notes = function(opts)
	opts = opts or {}
	local capytrace = require("capytrace")
	local save_path = capytrace.get_status().save_path or require("capytrace.config").get().save_path

	builtin.live_grep(vim.tbl_extend("force", {
		prompt_title = "CapyTrace Notes",
		search_dirs = { save_path },
		glob_pattern = "*.md",
		default_text = "Note",
	}, opts))
end

return telescope.register_extension({
	exports = {
		sessions = extension.sessions,
		notes = extension.notes,
	},
})
