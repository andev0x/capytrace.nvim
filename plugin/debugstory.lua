-- Plugin entry point
if vim.g.loaded_debugstory then
  return
end
vim.g.loaded_debugstory = true

-- Only load if Neovim version is compatible
if vim.fn.has('nvim-0.7') == 0 then
  vim.notify('debugstory.nvim requires Neovim 0.7+', vim.log.levels.ERROR)
  return
end

-- Set up the plugin
require('debugstory').setup()