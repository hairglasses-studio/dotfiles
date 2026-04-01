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

  -- Full header lines (revealed progressively)
  local header_lines = {
    [[                                                          ]],
    [[  ███╗   ██╗███████╗ ██████╗ ██╗   ██╗██╗███╗   ███╗    ]],
    [[  ████╗  ██║██╔════╝██╔═══██╗██║   ██║██║████╗ ████║    ]],
    [[  ██╔██╗ ██║█████╗  ██║   ██║██║   ██║██║██╔████╔██║    ]],
    [[  ██║╚██╗██║██╔══╝  ██║   ██║╚██╗ ██╔╝██║██║╚██╔╝██║    ]],
    [[  ██║ ╚████║███████╗╚██████╔╝ ╚████╔╝ ██║██║ ╚═╝ ██║    ]],
    [[  ╚═╝  ╚═══╝╚══════╝ ╚═════╝   ╚═══╝  ╚═╝╚═╝     ╚═╝    ]],
    [[                                                          ]],
    [[    ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄    ]],
    [[    █ C Y B E R N E T   //   S N A Z Z Y   v 3 . 0 █    ]],
    [[    ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀    ]],
  }

  -- Start with blank lines, progressively fill them in
  dashboard.section.header.val = {}
  for i = 1, #header_lines do
    dashboard.section.header.val[i] = ''
  end

  -- Gradient highlights: magenta → purple → cyan → green
  local header_colors = {
    'AlphaH1', 'AlphaH2', 'AlphaH3', 'AlphaH4',
    'AlphaH5', 'AlphaH6', 'AlphaH7', 'AlphaH8',
    'AlphaH9', 'AlphaH10', 'AlphaH11',
  }
  local header_hex = {
    '#ff6ac1', '#ff6ac1', '#9a8eef', '#57c7ff',
    '#57c7ff', '#9aedfe', '#5af78e', '#686868',
    '#57c7ff', '#5af78e', '#57c7ff',
  }
  -- Glitch/scramble highlight (used during reveal)
  vim.api.nvim_set_hl(0, 'AlphaGlitch', { fg = '#686868' })
  for i, name in ipairs(header_colors) do
    vim.api.nvim_set_hl(0, name, { fg = header_hex[i] })
  end

  local hl = {}
  for i = 1, #header_lines do
    hl[i] = { { 'AlphaGlitch', 0, -1 } }
  end
  dashboard.section.header.opts.hl = hl

  -- Boot status line (shown below header, above buttons)
  local boot_section = { type = 'text', val = '', opts = { hl = 'AlphaH4', position = 'center' } }

  -- Buttons
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

  local fortune_handle = io.popen('fortune -s 2>/dev/null')
  local fortune_text = fortune_handle and fortune_handle:read('*a') or ''
  if fortune_handle then fortune_handle:close() end
  dashboard.section.footer.val = fortune_text ~= '' and fortune_text or '  [ C Y B E R N E T   A C T I V E ]'
  dashboard.section.footer.opts.hl = 'AlphaH8'

  -- Insert boot status between header and buttons
  dashboard.config.layout = {
    { type = 'padding', val = 2 },
    dashboard.section.header,
    { type = 'padding', val = 1 },
    boot_section,
    { type = 'padding', val = 2 },
    dashboard.section.buttons,
    { type = 'padding', val = 1 },
    dashboard.section.footer,
  }

  alpha.setup(dashboard.config)

  -- ── Typewriter boot animation ──────────────
  -- Scramble characters for the glitch phase
  local glitch_chars = '░▒▓█▀▄╗╔╚╝━─│┃┌┐└┘├┤┬┴┼'

  local function scramble_line(line)
    local result = {}
    for i = 1, #line do
      local c = line:sub(i, i)
      if c == ' ' then
        result[i] = ' '
      else
        local gi = math.random(1, #glitch_chars)
        result[i] = glitch_chars:sub(gi, gi)
      end
    end
    return table.concat(result)
  end

  local function run_boot_animation()
    local buf = vim.api.nvim_get_current_buf()
    local ft = vim.bo[buf].filetype
    if ft ~= 'alpha' then return end

    -- Phase 1: Glitch-fill each line (fast, 30ms per line)
    for i, line in ipairs(header_lines) do
      vim.defer_fn(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end
        dashboard.section.header.val[i] = scramble_line(line)
        pcall(function() require('alpha').redraw() end)
      end, i * 30)
    end

    -- Phase 2: Resolve glitched lines to real text (40ms per line, after phase 1)
    local phase1_end = #header_lines * 30 + 80
    for i, line in ipairs(header_lines) do
      vim.defer_fn(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end
        dashboard.section.header.val[i] = line
        -- Switch highlight from glitch to final color
        dashboard.section.header.opts.hl[i] = { { header_colors[i], 0, -1 } }
        pcall(function() require('alpha').redraw() end)
      end, phase1_end + i * 40)
    end

    -- Phase 3: Boot status messages
    local phase2_end = phase1_end + #header_lines * 40 + 100
    local boot_msgs = {
      { msg = '> INITIALIZING NEURAL INTERFACE ...', hl = 'AlphaH4' },
      { msg = '> LOADING PLUGINS .................. OK', hl = 'AlphaH7' },
      { msg = '> SYNTAX ENGINES ................... ARMED', hl = 'AlphaH5' },
      { msg = '> SYSTEM READY // AWAITING INPUT', hl = 'AlphaH10' },
    }
    for i, bm in ipairs(boot_msgs) do
      vim.defer_fn(function()
        if not vim.api.nvim_buf_is_valid(buf) then return end
        boot_section.val = bm.msg
        boot_section.opts.hl = bm.hl
        pcall(function() require('alpha').redraw() end)
      end, phase2_end + i * 200)
    end
  end

  -- Trigger animation when alpha buffer opens
  vim.api.nvim_create_autocmd('User', {
    pattern = 'AlphaReady',
    once = true,
    callback = function()
      vim.defer_fn(run_boot_animation, 50)
    end,
  })
end
