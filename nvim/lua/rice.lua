-- rice.lua: Snazzy-themed Lua plugin configs

-- ── Treesitter ──────────────────────────────
local ts_ok, ts = pcall(require, 'nvim-treesitter.configs')
if ts_ok then
  ts.setup({
    ensure_installed = {
      'go', 'gomod', 'gosum', 'javascript', 'typescript', 'python',
      'lua', 'bash', 'json', 'yaml', 'toml', 'markdown', 'markdown_inline',
      'dockerfile', 'hcl', 'terraform', 'proto', 'html', 'css', 'vim', 'vimdoc',
    },
    highlight = { enable = true },
    indent = { enable = true },
  })
end

-- ── Indent Blankline ────────────────────────
local ibl_ok, ibl = pcall(require, 'ibl')
if ibl_ok then
  ibl.setup({
    indent = {
      char = '│',
      highlight = { 'IblIndent' },
    },
    scope = {
      enabled = true,
      highlight = { 'IblScope' },
      show_start = false,
      show_end = false,
    },
  })
  vim.api.nvim_set_hl(0, 'IblIndent', { fg = '#2a2a2a' })
  vim.api.nvim_set_hl(0, 'IblScope', { fg = '#57c7ff' })
end

-- ── Colorizer ───────────────────────────────
local col_ok, colorizer = pcall(require, 'colorizer')
if col_ok then
  colorizer.setup({ '*' }, {
    RGB = true,
    RRGGBB = true,
    names = false,
    RRGGBBAA = true,
    css = true,
    css_fn = true,
  })
end

-- ── Alpha Dashboard ─────────────────────────
local alpha_ok, alpha = pcall(require, 'alpha')
if alpha_ok then
  local dashboard = require('alpha.themes.dashboard')

  dashboard.section.header.val = {
    [[                                              ]],
    [[    ██████╗ ██╗ ██████╗███████╗██████╗        ]],
    [[    ██╔══██╗██║██╔════╝██╔════╝██╔══██╗       ]],
    [[    ██████╔╝██║██║     █████╗  ██║  ██║       ]],
    [[    ██╔══██╗██║██║     ██╔══╝  ██║  ██║       ]],
    [[    ██║  ██║██║╚██████╗███████╗██████╔╝       ]],
    [[    ╚═╝  ╚═╝╚═╝ ╚═════╝╚══════╝╚═════╝        ]],
    [[                                              ]],
    [[          ╔═══════════════════════╗            ]],
    [[          ║   S N A Z Z Y  v i m ║            ]],
    [[          ╚═══════════════════════╝            ]],
  }

  -- Gradient highlights: magenta → blue → green
  local header_colors = {
    'AlphaH1', 'AlphaH2', 'AlphaH3', 'AlphaH4',
    'AlphaH5', 'AlphaH6', 'AlphaH7', 'AlphaH8',
    'AlphaH9', 'AlphaH10', 'AlphaH11',
  }
  local header_hex = {
    '#ff6ac1', '#ff6ac1', '#9a8eef', '#57c7ff',
    '#57c7ff', '#5af78e', '#5af78e', '#686868',
    '#57c7ff', '#57c7ff', '#686868',
  }
  for i, name in ipairs(header_colors) do
    vim.api.nvim_set_hl(0, name, { fg = header_hex[i] })
  end

  local hl = {}
  for i = 1, #dashboard.section.header.val do
    hl[i] = { { header_colors[i], 0, -1 } }
  end
  dashboard.section.header.opts.hl = hl

  dashboard.section.buttons.val = {
    dashboard.button('f', '  Find file',       ':Files<CR>'),
    dashboard.button('r', '  Recent files',    ':History<CR>'),
    dashboard.button('g', '  Grep text',       ':Rg<CR>'),
    dashboard.button('n', '  New file',        ':enew<CR>'),
    dashboard.button('c', '  Configuration',   ':edit $MYVIMRC<CR>'),
    dashboard.button('q', '  Quit',            ':qa<CR>'),
  }

  for _, button in ipairs(dashboard.section.buttons.val) do
    button.opts.hl = 'AlphaH4'
    button.opts.hl_shortcut = 'AlphaH1'
  end

  dashboard.section.footer.val = '  Snazzy on Black'
  dashboard.section.footer.opts.hl = 'AlphaH8'

  alpha.setup(dashboard.config)
end
