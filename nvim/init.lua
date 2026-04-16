-- Cross-platform Neovim configuration
-- Migrated from init.vim to lazy.nvim

-- ── Leader key (must be before lazy) ────────────
vim.g.mapleader = ' '
vim.g.maplocalleader = '\\'

-- ── Bootstrap lazy.nvim ─────────────────────────
local lazypath = vim.fn.stdpath('data') .. '/lazy/lazy.nvim'
if not vim.loop.fs_stat(lazypath) then
  vim.fn.system({
    'git', 'clone', '--filter=blob:none',
    'https://github.com/folke/lazy.nvim.git',
    '--branch=stable', lazypath,
  })
end
vim.opt.rtp:prepend(lazypath)

-- ── Plugin specs ────────────────────────────────
require('lazy').setup({
  -- Essential editing
  'tpope/vim-sensible',
  'tpope/vim-surround',
  'tpope/vim-commentary',
  'jiangmiao/auto-pairs',
  { 'machakann/vim-highlightedyank', event = 'TextYankPost' },

  -- File navigation
  { 'junegunn/fzf', build = function() vim.fn['fzf#install']() end },
  'junegunn/fzf.vim',
  {
    'preservim/nerdtree',
    keys = {
      { '<leader>n', '<cmd>NERDTreeToggle<cr>' },
      { '<leader>nf', '<cmd>NERDTreeFind<cr>' },
    },
    config = function()
      vim.g.NERDTreeShowHidden = 1
      vim.g.NERDTreeMinimalUI = 1
      vim.g.NERDTreeIgnore = { '\\.pyc$', '\\.pyo$', '\\.class$', '\\.o$', '\\~$' }
    end,
  },
  'ryanoasis/vim-devicons',
  'nvim-tree/nvim-web-devicons',

  -- Git
  'tpope/vim-fugitive',
  'tpope/vim-rhubarb',
  { 'airblade/vim-gitgutter', config = function()
    vim.g.gitgutter_enabled = 1
    vim.g.gitgutter_map_keys = 0
  end },

  -- LSP / Completion
  { 'neoclide/coc.nvim', branch = 'release' },

  -- Color schemes
  { 'connorholyday/vim-snazzy', priority = 1000 },
  { 'scottmckendry/cyberdream.nvim', lazy = true },
  { 'hyperb1iss/silkcircuit-nvim', lazy = true },

  -- Status line
  { 'vim-airline/vim-airline', config = function()
    vim.g.airline_powerline_fonts = 1
    vim.g['airline#extensions#tabline#enabled'] = 1
    vim.g['airline#extensions#tabline#formatter'] = 'unique_tail'
    vim.g.airline_theme = 'base16_snazzy'
  end },
  'vim-airline/vim-airline-themes',

  -- Markdown
  { 'plasticboy/vim-markdown', ft = 'markdown', config = function()
    vim.g.vim_markdown_folding_disabled = 1
    vim.g.vim_markdown_frontmatter = 1
    vim.g.vim_markdown_toml_frontmatter = 1
    vim.g.vim_markdown_json_frontmatter = 1
    vim.g.vim_markdown_new_list_item_indent = 2
  end },
  { 'iamcco/markdown-preview.nvim', ft = 'markdown', build = 'cd app && yarn install' },

  -- Utilities
  { 'mbbill/undotree', keys = { { '<leader>u', '<cmd>UndotreeToggle<cr>' } },
    config = function()
      vim.g.undotree_WindowLayout = 2
      vim.g.undotree_ShortIndicators = 1
    end },
  'lambdalisue/suda.vim',
  'christoomey/vim-tmux-navigator',
  'editorconfig/editorconfig-vim',

  -- Kitty integration
  { 'mikesmithgh/kitty-scrollback.nvim',
    lazy = true,
    cmd = { 'KittyScrollbackGenerateKittens', 'KittyScrollbackGenerateCommandLineEditing' },
    event = { 'User KittyScrollbackLaunch' },
    config = function() require('kitty-scrollback').setup() end,
  },

  -- Modern enhancements
  { 'goolord/alpha-nvim', event = 'VimEnter' },
  { 'nvim-treesitter/nvim-treesitter', build = ':TSUpdate' },
  { 'lukas-reineke/indent-blankline.nvim', main = 'ibl' },
  { 'norcalli/nvim-colorizer.lua', event = 'BufReadPre' },
  'nvim-lua/plenary.nvim',
}, {
  ui = {
    border = 'rounded',
    icons = { loaded = '●', not_loaded = '○' },
  },
  performance = {
    rtp = {
      disabled_plugins = { 'gzip', 'tarPlugin', 'tohtml', 'tutor', 'zipPlugin' },
    },
  },
})

-- ── Options ─────────────────────────────────────
local o = vim.opt

-- Display
o.number = true
o.relativenumber = true
o.cursorline = true
o.colorcolumn = '80,120'
o.signcolumn = 'yes'
o.termguicolors = true
o.background = 'dark'
o.list = true
o.listchars = { tab = '›\\ ', trail = '•', extends = '#', nbsp = '·', eol = '¬' }
o.title = true
o.titlestring = '%t%( %M%)%( (%{expand("%:~:.:h")})%)%( %a%)'

-- Search
o.ignorecase = true
o.smartcase = true
o.incsearch = true
o.hlsearch = true

-- Indentation
o.expandtab = true
o.tabstop = 4
o.shiftwidth = 4
o.softtabstop = 4
o.autoindent = true
o.smartindent = true

-- Editor behavior
o.hidden = true
o.wrap = true
o.scrolloff = 5
o.sidescrolloff = 5
o.mouse = 'a'
o.clipboard = 'unnamed,unnamedplus'
o.updatetime = 300
o.timeoutlen = 500
o.undofile = true
o.backup = true
o.writebackup = true
o.swapfile = true
o.backupdir = vim.fn.expand('~/.local/share/nvim/backup//')
o.undodir = vim.fn.expand('~/.local/share/nvim/undo//')
o.directory = vim.fn.expand('~/.local/share/nvim/swap//')

-- Folding
o.foldmethod = 'indent'
o.foldlevelstart = 99
o.foldnestmax = 10

-- Splits
o.splitright = true
o.splitbelow = true

-- Wildmenu
o.wildmenu = true
o.wildmode = 'longest:full,full'
o.wildignore:append({ '*.o', '*.obj', '*.pyc', '*.class', '*/.git/*', '*/node_modules/*', '*/__pycache__/*' })

-- Performance
o.lazyredraw = true
o.regexpengine = 1
o.synmaxcol = 200

-- Create backup dirs
vim.fn.mkdir(vim.fn.expand('~/.local/share/nvim/backup'), 'p')
vim.fn.mkdir(vim.fn.expand('~/.local/share/nvim/undo'), 'p')
vim.fn.mkdir(vim.fn.expand('~/.local/share/nvim/swap'), 'p')

-- ── Colorscheme ─────────────────────────────────
vim.g.SnazzyTransparent = 1
pcall(vim.cmd.colorscheme, 'snazzy')

-- Theme toggle: <leader>ct cycles snazzy → cyberdream → silkcircuit
vim.keymap.set('n', '<leader>ct', function()
  local themes = { 'snazzy', 'cyberdream', 'silkcircuit' }
  local cur = vim.g.colors_name or 'snazzy'
  for i, t in ipairs(themes) do
    if t == cur then
      local next_theme = themes[i % #themes + 1]
      vim.cmd('colorscheme ' .. next_theme)
      print('Theme: ' .. next_theme)
      break
    end
  end
end)

-- ── Key Mappings ────────────────────────────────
local map = vim.keymap.set

-- Quick save/quit
map('n', '<leader>w', '<cmd>w<cr>')
map('n', '<leader>q', '<cmd>q<cr>')
map('n', '<leader>x', '<cmd>x<cr>')

-- Buffers
map('n', '<leader>bn', '<cmd>bnext<cr>')
map('n', '<leader>bp', '<cmd>bprevious<cr>')
map('n', '<leader>bd', '<cmd>bdelete<cr>')

-- Window resize
map('n', '<leader>=', '<C-w>=')
map('n', '<leader>+', '<cmd>resize +5<cr>')
map('n', '<leader>-', '<cmd>resize -5<cr>')
map('n', '<leader><', '<cmd>vertical resize -5<cr>')
map('n', '<leader>>', '<cmd>vertical resize +5<cr>')

-- Tabs
map('n', '<leader>tn', '<cmd>tabnew<cr>')
map('n', '<leader>tc', '<cmd>tabclose<cr>')
map('n', '<leader>th', '<cmd>tabprevious<cr>')
map('n', '<leader>tl', '<cmd>tabnext<cr>')

-- Clear search
map('n', '<Esc>', '<cmd>nohlsearch<cr>')

-- Config editing
map('n', '<leader>ev', '<cmd>edit $MYVIMRC<cr>')

-- FZF
map('n', '<leader>f', '<cmd>Files<cr>')
map('n', '<leader>b', '<cmd>Buffers<cr>')
map('n', '<leader>l', '<cmd>Lines<cr>')
map('n', '<leader>h', '<cmd>History<cr>')
map('n', '<leader>:', '<cmd>Commands<cr>')
map('n', '<leader>/', '<cmd>Rg<cr>')

-- Git (fugitive)
map('n', '<leader>gs', '<cmd>Git<cr>')
map('n', '<leader>ga', '<cmd>Gwrite<cr>')
map('n', '<leader>gc', '<cmd>Git commit<cr>')
map('n', '<leader>gp', '<cmd>Git push<cr>')
map('n', '<leader>gd', '<cmd>Gdiffsplit<cr>')
map('n', '<leader>gb', '<cmd>Git blame<cr>')
map('n', '<leader>gl', '<cmd>Git log --oneline --graph --decorate --all<cr>')

-- GitGutter
map('n', '<leader>gg', '<cmd>GitGutterToggle<cr>')
map('n', '<leader>gh', '<cmd>GitGutterLineHighlightsToggle<cr>')
map('n', ']h', '<cmd>GitGutterNextHunk<cr>')
map('n', '[h', '<cmd>GitGutterPrevHunk<cr>')

-- Markdown preview
map('n', '<leader>mp', '<cmd>MarkdownPreview<cr>')
map('n', '<leader>ms', '<cmd>MarkdownPreviewStop<cr>')

-- TODO search
map('n', '<leader>todo', function() vim.cmd("Rg 'TODO|FIXME|NOTE|HACK|BUG'") end)

-- Suda
vim.cmd([[cnoreabbrev w!! w suda://%]])

-- ── FZF config ──────────────────────────────────
vim.g.fzf_layout = { down = '~40%' }
vim.g.fzf_preview_window = { 'right:50%:hidden', 'ctrl-/' }

vim.cmd([[
command! -bang -nargs=? -complete=dir Files
    \ call fzf#vim#files(<q-args>, fzf#vim#with_preview(), <bang>0)

command! -bang -nargs=* Rg
    \ call fzf#vim#grep(
    \   'rg --column --line-number --no-heading --color=always --smart-case '.shellescape(<q-args>), 1,
    \   fzf#vim#with_preview(), <bang>0)
]])

-- ── CoC config ──────────────────────────────────
vim.cmd([[
inoremap <silent><expr> <TAB>
      \ pumvisible() ? "\<C-n>" :
      \ <SID>check_back_space() ? "\<TAB>" :
      \ coc#refresh()
inoremap <expr><S-TAB> pumvisible() ? "\<C-p>" : "\<C-h>"

function! s:check_back_space() abort
  let col = col('.') - 1
  return !col || getline('.')[col - 1]  =~# '\s'
endfunction

inoremap <silent><expr> <c-space> coc#refresh()
inoremap <silent><expr> <cr> pumvisible() ? coc#_select_confirm()
                              \: "\<C-g>u\<CR>\<c-r>=coc#on_enter()\<CR>"

nmap <silent> gd <Plug>(coc-definition)
nmap <silent> gy <Plug>(coc-type-definition)
nmap <silent> gi <Plug>(coc-implementation)
nmap <silent> gr <Plug>(coc-references)
nmap <leader>rn <Plug>(coc-rename)
xmap <leader>F  <Plug>(coc-format-selected)
nmap <leader>F  <Plug>(coc-format-selected)
]])

-- K for documentation
map('n', 'K', function()
  local ft = vim.bo.filetype
  if ft == 'vim' or ft == 'help' then
    vim.cmd('h ' .. vim.fn.expand('<cword>'))
  elseif vim.fn['coc#rpc#ready']() then
    vim.fn.CocActionAsync('doHover')
  else
    vim.cmd('!' .. vim.o.keywordprg .. ' ' .. vim.fn.expand('<cword>'))
  end
end)

-- ── Filetype settings ───────────────────────────
vim.api.nvim_create_augroup('FiletypeSettings', { clear = true })

local ft_settings = {
  { 'python',                   { tabstop = 4, shiftwidth = 4, textwidth = 79 } },
  { 'javascript,typescript,json', { tabstop = 2, shiftwidth = 2 } },
  { 'html,css,scss',            { tabstop = 2, shiftwidth = 2 } },
  { 'sh,bash,zsh',              { tabstop = 4, shiftwidth = 4 } },
  { 'yaml',                     { tabstop = 2, shiftwidth = 2 } },
  { 'markdown',                 { wrap = true, textwidth = 80, spell = true, conceallevel = 2 } },
}

for _, entry in ipairs(ft_settings) do
  vim.api.nvim_create_autocmd('FileType', {
    group = 'FiletypeSettings',
    pattern = entry[1],
    callback = function()
      for k, v in pairs(entry[2]) do vim.opt_local[k] = v end
    end,
  })
end

-- Strip trailing whitespace
vim.api.nvim_create_autocmd('BufWritePre', {
  group = 'FiletypeSettings',
  pattern = '*',
  callback = function()
    local pos = vim.api.nvim_win_get_cursor(0)
    vim.cmd([[%s/\s\+$//e]])
    vim.api.nvim_win_set_cursor(0, pos)
  end,
})

-- Return to last edit position
vim.api.nvim_create_autocmd('BufReadPost', {
  group = 'FiletypeSettings',
  pattern = '*',
  callback = function()
    local mark = vim.api.nvim_buf_get_mark(0, '"')
    if mark[1] > 1 and mark[1] <= vim.api.nvim_buf_line_count(0) then
      vim.api.nvim_win_set_cursor(0, mark)
    end
  end,
})

-- ── Lua plugin configs (rice.lua) ───────────────
require('rice')
