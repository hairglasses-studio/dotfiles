" Cross-platform Neovim configuration optimized for LLM agents
" Compatible with both macOS and Linux

" Ensure compatibility with both Vim and Neovim
if !has('nvim')
    set nocompatible
endif

" Plugin management using vim-plug
call plug#begin('~/.local/share/nvim/site/pack/packer/start')

" Essential editing plugins
Plug 'tpope/vim-sensible'           " Sensible defaults
Plug 'tpope/vim-surround'           " Surround text objects
Plug 'tpope/vim-commentary'         " Comment/uncomment code
Plug 'jiangmiao/auto-pairs'         " Auto-close brackets, quotes
Plug 'machakann/vim-highlightedyank' " Highlight yanked text

" File navigation and search
Plug 'junegunn/fzf', { 'do': { -> fzf#install() } }
Plug 'junegunn/fzf.vim'            " FZF integration
Plug 'preservim/nerdtree'           " File explorer
Plug 'ryanoasis/vim-devicons'       " File type icons

" Git integration
Plug 'tpope/vim-fugitive'           " Git wrapper
Plug 'airblade/vim-gitgutter'       " Git diff in gutter
Plug 'tpope/vim-rhubarb'            " GitHub integration

" Language support and completion
Plug 'neoclide/coc.nvim', {'branch': 'release'} " Completion engine
Plug 'sheerun/vim-polyglot'         " Language pack

" Syntax highlighting and themes
Plug 'connorholyday/vim-snazzy'     " Snazzy color scheme
Plug 'vim-airline/vim-airline'      " Status line
Plug 'vim-airline/vim-airline-themes' " Airline themes

" Markdown and documentation
Plug 'plasticboy/vim-markdown'      " Markdown support
Plug 'iamcco/markdown-preview.nvim', { 'do': 'cd app && yarn install' }

" Productivity and utilities
Plug 'mbbill/undotree'              " Undo history visualizer
Plug 'lambdalisue/suda.vim'         " Sudo support
Plug 'christoomey/vim-tmux-navigator' " Tmux navigation
Plug 'editorconfig/editorconfig-vim' " EditorConfig support

call plug#end()

" ===============================
" Basic Settings
" ===============================

" Enable filetype detection and plugins
filetype plugin indent on
syntax enable

" Set encoding
set encoding=utf-8
set fileencoding=utf-8

" Line numbers and visual aids
set number relativenumber           " Hybrid line numbers
set cursorline                      " Highlight current line
set colorcolumn=80,120              " Rulers at 80 and 120 characters
set signcolumn=yes                  " Always show sign column

" Search settings
set ignorecase smartcase            " Smart case searching
set incsearch                       " Incremental search
set hlsearch                        " Highlight search results

" Indentation and formatting
set expandtab                       " Use spaces instead of tabs
set tabstop=4                       " Tab width
set shiftwidth=4                    " Indentation width
set softtabstop=4                   " Soft tab width
set autoindent                      " Auto-indent new lines
set smartindent                     " Smart indentation

" Editor behavior
set hidden                          " Allow hidden buffers
set wrap                            " Wrap long lines
set scrolloff=5                     " Keep 5 lines visible when scrolling
set sidescrolloff=5                 " Keep 5 columns visible when scrolling
set mouse=a                         " Enable mouse support
set clipboard^=unnamed,unnamedplus  " Use system clipboard
set updatetime=300                  " Faster completion
set timeoutlen=500                  " Faster key sequences
set undofile                        " Persistent undo
set backup                          " Keep backups
set writebackup                     " Write backups
set swapfile                        " Use swap files

" Set backup and undo directories
set backupdir=~/.local/share/nvim/backup//
set undodir=~/.local/share/nvim/undo//
set directory=~/.local/share/nvim/swap//

" Create directories if they don't exist
silent !mkdir -p ~/.local/share/nvim/backup
silent !mkdir -p ~/.local/share/nvim/undo  
silent !mkdir -p ~/.local/share/nvim/swap

" Folding
set foldmethod=indent
set foldlevelstart=99               " Start with all folds open
set foldnestmax=10                  " Maximum fold nesting

" Window splitting
set splitright                      " Split to the right
set splitbelow                      " Split below

" Wildmenu settings
set wildmenu
set wildmode=longest:full,full
set wildignore+=*.o,*.obj,*.pyc,*.class
set wildignore+=*/.git/*,*/node_modules/*,*/__pycache__/*

" ===============================
" Visual Settings
" ===============================

" Color scheme
set termguicolors                   " Enable true color
set background=dark
let g:SnazzyTransparent = 1
colorscheme snazzy

" Airline configuration
let g:airline_powerline_fonts = 1
let g:airline#extensions#tabline#enabled = 1
let g:airline#extensions#tabline#formatter = 'unique_tail'
let g:airline_theme = 'base16_snazzy'

" Show whitespace characters
set list
set listchars=tab:›\ ,trail:•,extends:#,nbsp:·,eol:¬

" ===============================
" Key Mappings
" ===============================

" Set leader key
let mapleader = "\<Space>"
let maplocalleader = "\\"

" Quick save and quit
nnoremap <leader>w :w<CR>
nnoremap <leader>q :q<CR>
nnoremap <leader>x :x<CR>

" Buffer navigation
nnoremap <leader>bn :bnext<CR>
nnoremap <leader>bp :bprevious<CR>
nnoremap <leader>bd :bdelete<CR>

" Window navigation (with tmux integration)
nnoremap <C-h> <C-w>h
nnoremap <C-j> <C-w>j
nnoremap <C-k> <C-w>k
nnoremap <C-l> <C-w>l

" Window resizing
nnoremap <leader>= <C-w>=
nnoremap <leader>+ :resize +5<CR>
nnoremap <leader>- :resize -5<CR>
nnoremap <leader>< :vertical resize -5<CR>
nnoremap <leader>> :vertical resize +5<CR>

" Tab navigation
nnoremap <leader>tn :tabnew<CR>
nnoremap <leader>tc :tabclose<CR>
nnoremap <leader>th :tabprevious<CR>
nnoremap <leader>tl :tabnext<CR>

" Clear search highlighting
nnoremap <leader>/ :nohlsearch<CR>

" Quick editing of configuration
nnoremap <leader>ev :edit $MYVIMRC<CR>
nnoremap <leader>sv :source $MYVIMRC<CR>

" ===============================
" Plugin Configuration
" ===============================

" NERDTree
nnoremap <leader>n :NERDTreeToggle<CR>
nnoremap <leader>nf :NERDTreeFind<CR>
let g:NERDTreeShowHidden = 1
let g:NERDTreeMinimalUI = 1
let g:NERDTreeIgnore = ['\.pyc$', '\.pyo$', '\.rbc$', '\.rbo$', '\.class$', '\.o$', '\~$']

" FZF
nnoremap <leader>f :Files<CR>
nnoremap <leader>F :Files<CR>
nnoremap <leader>b :Buffers<CR>
nnoremap <leader>l :Lines<CR>
nnoremap <leader>h :History<CR>
nnoremap <leader>: :Commands<CR>
nnoremap <leader>/ :Rg<CR>

" FZF window settings
let g:fzf_layout = { 'down': '~40%' }
let g:fzf_preview_window = ['right:50%:hidden', 'ctrl-/']

" FZF commands with preview
command! -bang -nargs=? -complete=dir Files
    \ call fzf#vim#files(<q-args>, fzf#vim#with_preview(), <bang>0)

command! -bang -nargs=* Rg
    \ call fzf#vim#grep(
    \   'rg --column --line-number --no-heading --color=always --smart-case '.shellescape(<q-args>), 1,
    \   fzf#vim#with_preview(), <bang>0)

" Git (Fugitive)
nnoremap <leader>gs :Git<CR>
nnoremap <leader>ga :Gwrite<CR>
nnoremap <leader>gc :Git commit<CR>
nnoremap <leader>gp :Git push<CR>
nnoremap <leader>gd :Gdiffsplit<CR>
nnoremap <leader>gb :Git blame<CR>
nnoremap <leader>gl :Git log --oneline --graph --decorate --all<CR>

" GitGutter
let g:gitgutter_enabled = 1
let g:gitgutter_map_keys = 0
nnoremap <leader>gg :GitGutterToggle<CR>
nnoremap <leader>gh :GitGutterLineHighlightsToggle<CR>
nnoremap ]h :GitGutterNextHunk<CR>
nnoremap [h :GitGutterPrevHunk<CR>

" UndoTree
nnoremap <leader>u :UndotreeToggle<CR>
let g:undotree_WindowLayout = 2
let g:undotree_ShortIndicators = 1

" Markdown
let g:vim_markdown_folding_disabled = 1
let g:vim_markdown_frontmatter = 1
let g:vim_markdown_toml_frontmatter = 1
let g:vim_markdown_json_frontmatter = 1
let g:vim_markdown_new_list_item_indent = 2

" Markdown Preview
nnoremap <leader>mp :MarkdownPreview<CR>
nnoremap <leader>ms :MarkdownPreviewStop<CR>

" Suda (sudo support)
cnoreabbrev w!! w suda://%

" CoC configuration
" Use tab for trigger completion
inoremap <silent><expr> <TAB>
      \ pumvisible() ? "\<C-n>" :
      \ <SID>check_back_space() ? "\<TAB>" :
      \ coc#refresh()
inoremap <expr><S-TAB> pumvisible() ? "\<C-p>" : "\<C-h>"

function! s:check_back_space() abort
  let col = col('.') - 1
  return !col || getline('.')[col - 1]  =~# '\s'
endfunction

" Use <c-space> to trigger completion
inoremap <silent><expr> <c-space> coc#refresh()

" Use <cr> to confirm completion
inoremap <silent><expr> <cr> pumvisible() ? coc#_select_confirm()
                              \: "\<C-g>u\<CR>\<c-r>=coc#on_enter()\<CR>"

" GoTo code navigation
nmap <silent> gd <Plug>(coc-definition)
nmap <silent> gy <Plug>(coc-type-definition)
nmap <silent> gi <Plug>(coc-implementation)
nmap <silent> gr <Plug>(coc-references)

" Use K to show documentation
nnoremap <silent> K :call <SID>show_documentation()<CR>

function! s:show_documentation()
  if (index(['vim','help'], &filetype) >= 0)
    execute 'h '.expand('<cword>')
  elseif (coc#rpc#ready())
    call CocActionAsync('doHover')
  else
    execute '!' . &keywordprg . " " . expand('<cword>')
  endif
endfunction

" Symbol renaming
nmap <leader>rn <Plug>(coc-rename)

" Formatting selected code
xmap <leader>F  <Plug>(coc-format-selected)
nmap <leader>F  <Plug>(coc-format-selected)

" ===============================
" LLM-Friendly Features
" ===============================

" Function to show current context for LLM agents
function! LLMContext()
    echo "=== Current Buffer Info ==="
    echo "File: " . expand('%:p')
    echo "Filetype: " . &filetype
    echo "Lines: " . line('$')
    echo "Current position: " . line('.') . ":" . col('.')
    echo "=== Git Status ==="
    if exists(':Git')
        Git status --porcelain
    endif
    echo "=== Recent Changes ==="
    if exists(':GitGutter')
        GitGutterAll
    endif
endfunction
nnoremap <leader>llm :call LLMContext()<CR>

" Function to create a project outline
function! ProjectOutline()
    if executable('find')
        echo "=== Project Structure ==="
        !find . -type f -name "*.py" -o -name "*.js" -o -name "*.ts" -o -name "*.sh" -o -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" | head -20
    endif
    if executable('git')
        echo "=== Git Files ==="
        !git ls-files | head -20
    endif
endfunction
nnoremap <leader>po :call ProjectOutline()<CR>

" Quick function/class navigation
nnoremap <leader>ff :Lines function\|class\|def<CR>

" Show all TODO/FIXME/NOTE comments
nnoremap <leader>todo :Rg 'TODO\|FIXME\|NOTE\|HACK\|BUG'<CR>

" ===============================
" File Type Specific Settings
" ===============================

augroup FiletypeSettings
    autocmd!
    
    " Python
    autocmd FileType python setlocal tabstop=4 shiftwidth=4 softtabstop=4
    autocmd FileType python setlocal textwidth=79
    autocmd FileType python nnoremap <buffer> <leader>r :!python %<CR>
    
    " JavaScript/TypeScript
    autocmd FileType javascript,typescript,json setlocal tabstop=2 shiftwidth=2 softtabstop=2
    autocmd FileType javascript nnoremap <buffer> <leader>r :!node %<CR>
    
    " HTML/CSS
    autocmd FileType html,css,scss setlocal tabstop=2 shiftwidth=2 softtabstop=2
    
    " Shell scripts
    autocmd FileType sh,bash,zsh setlocal tabstop=4 shiftwidth=4 softtabstop=4
    autocmd FileType sh nnoremap <buffer> <leader>r :!bash %<CR>
    
    " Markdown
    autocmd FileType markdown setlocal wrap textwidth=80 spell
    autocmd FileType markdown setlocal conceallevel=2
    
    " YAML
    autocmd FileType yaml setlocal tabstop=2 shiftwidth=2 softtabstop=2
    
    " Remove trailing whitespace on save
    autocmd BufWritePre * :%s/\s\+$//e
    
    " Return to last edit position when opening files
    autocmd BufReadPost * if line("'\"") > 1 && line("'\"") <= line("$") | exe "normal! g'\"" | endif
augroup END

" ===============================
" Status Line Information
" ===============================

" Custom status line function for debugging
function! LLMStatusLine()
    let l:status = ''
    let l:status .= ' %f'                    " filename
    let l:status .= ' %m'                    " modified flag
    let l:status .= ' %='                    " right align
    let l:status .= ' %{&filetype}'          " filetype
    let l:status .= ' %l:%c'                 " line:column
    let l:status .= ' %p%%'                  " percentage
    return l:status
endfunction

" Load local configuration if it exists
if filereadable(expand('~/.config/nvim/local.vim'))
    source ~/.config/nvim/local.vim
endif

" ===============================
" Final Settings
" ===============================

" Ensure proper color support
if $TERM == 'xterm-256color'
    set t_Co=256
endif

" Enable syntax highlighting
if !exists("g:syntax_on")
    syntax enable
endif

" Set title for terminal
set title
set titlestring=%t%(\ %M%)%(\ (%{expand(\"%:~:.:h\")})%)%(\ %a%)

" Performance optimizations
set lazyredraw                      " Don't redraw while executing macros
set regexpengine=1                  " Use old regexp engine (faster)
set synmaxcol=200                   " Don't highlight very long lines

" Final message
echom "Neovim configured for cross-platform LLM agent use!"
