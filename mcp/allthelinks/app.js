'use strict';

// ── Default data (seed from eBay mini PC watchlist) ──

var DEFAULT_DATA = {
  settings: {
    title: 'allthelinks',
    columns: 1,
    showClock: true,
    showSearch: true,
    theme: 'snazzy'
  },
  groups: [{
    id: 'g_ebay001',
    title: 'Mini PC eBay Watch List',
    icon: '\u{1F5A5}\uFE0F',
    collapsed: false,
    links: [
      {
        id: 'l_ser7', label: 'Beelink SER7 \u2014 Ryzen 7 7840HS',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER7+7840HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=100&_udhi=350&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER7+7840HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_ItemCondition=3000&_udlo=100&_udhi=350&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Radeon 780M \u00b7 RDNA 3', color: '#bb86fc' },
          { text: 'DDR5 SO-DIMM', color: '#64b5f6' },
          { text: '$300\u2013380 w/ RAM', color: '#6be5a0' },
          { text: '\u26a0 buy w/ RAM included', color: '#ff8a65' }
        ],
        note: 'Best overall. GTX 1050 Ti\u20131650 class. 1080p med esports, 1080p low AAA w/ FSR.'
      },
      {
        id: 'l_um780', label: 'Minisforum UM780 XTX \u2014 Ryzen 7 7840HS',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Minisforum+UM780+XTX+7840HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=100&_udhi=380&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Minisforum+UM780+XTX+7840HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_ItemCondition=3000&_udlo=100&_udhi=380&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Radeon 780M \u00b7 RDNA 3', color: '#bb86fc' },
          { text: 'DDR5 SO-DIMM', color: '#64b5f6' },
          { text: '$300\u2013380 w/ RAM', color: '#6be5a0' },
          { text: 'USB4 + OCuLink for eGPU', color: '#ff8a65' }
        ],
        note: 'Same GPU as SER7. OCuLink is a real eGPU upgrade path later.'
      },
      {
        id: 'l_hvk', label: 'Intel NUC Hades Canyon NUC8i7HVK',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=NUC8i7HVK&_ex_kw=parts+repair+broken+motherboard+board&LH_BIN=1&LH_ItemCondition=3000&_udlo=100&_udhi=300&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=NUC8i7HVK&_ex_kw=parts+repair+broken+motherboard+board&LH_ItemCondition=3000&_udlo=100&_udhi=300&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Vega M GH \u00b7 4GB HBM2 dedicated', color: '#bb86fc' },
          { text: 'DDR4 SO-DIMM', color: '#64b5f6' },
          { text: '$270\u2013375 all-in', color: '#6be5a0' },
          { text: '\u2705 confirmed sub-$300', color: '#6be5a0' }
        ],
        note: 'Only dedicated-GPU NUC under $300. DDR4 = cheap RAM. 2018 CPU is the trade-off.'
      },
      {
        id: 'l_ser6', label: 'Beelink SER6 Pro \u2014 Ryzen 7 7735HS',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER6+Pro+7735HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=100&_udhi=300&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER6+Pro+7735HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_ItemCondition=3000&_udlo=100&_udhi=300&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Radeon 680M \u00b7 RDNA 2', color: '#bb86fc' },
          { text: 'DDR5 SO-DIMM', color: '#64b5f6' },
          { text: '$250\u2013350 w/ RAM', color: '#6be5a0' }
        ],
        note: '15\u201320% behind 780M. Sweet spot at $250\u2013280 configured.'
      },
      {
        id: 'l_ser5max', label: 'Beelink SER5 MAX \u2014 Ryzen 7 7735HS',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER5+MAX+7735HS&_ex_kw=parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=75&_udhi=300&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER5+MAX+7735HS&_ex_kw=parts+repair+broken&LH_ItemCondition=3000&_udlo=75&_udhi=300&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Radeon 680M \u00b7 RDNA 2', color: '#bb86fc' },
          { text: '24GB LPDDR5 soldered', color: '#64b5f6' },
          { text: '$200\u2013300', color: '#6be5a0' },
          { text: '\u2705 RAM crisis\u2013proof', color: '#6be5a0' }
        ],
        note: 'Soldered RAM = no separate purchase. 24GB is plenty for gaming. Best recession-proof pick.'
      },
      {
        id: 'l_a7', label: 'GEEKOM A7 \u2014 Ryzen 7 7840HS',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=GEEKOM+A7+7840HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=100&_udhi=350&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=GEEKOM+A7+7840HS&_ex_kw=barebones+barebone+parts+repair+broken&LH_ItemCondition=3000&_udlo=100&_udhi=350&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Radeon 780M \u00b7 RDNA 3', color: '#bb86fc' },
          { text: 'DDR5 SO-DIMM', color: '#64b5f6' },
          { text: '$300\u2013430 w/ RAM', color: '#6be5a0' },
          { text: '\u26a0 stretch budget', color: '#ff8a65' }
        ],
        note: 'Premium build. Same 780M. Typically priced above SER7 \u2014 need a deal.'
      },
      {
        id: 'l_um790', label: 'Minisforum UM790 Pro \u2014 Ryzen 9 7940HS',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Minisforum+UM790+Pro&_ex_kw=barebones+barebone+parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=150&_udhi=400&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Minisforum+UM790+Pro&_ex_kw=barebones+barebone+parts+repair+broken&LH_ItemCondition=3000&_udlo=150&_udhi=400&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Radeon 780M \u00b7 RDNA 3', color: '#bb86fc' },
          { text: 'DDR5 SO-DIMM', color: '#64b5f6' },
          { text: '$350\u2013450 w/ RAM', color: '#6be5a0' },
          { text: '\u274c over budget typical', color: '#ff8a65' }
        ],
        note: 'Ryzen 9 badge = marginal gains over 7840HS. Only worth it if priced same as SER7.'
      },
      {
        id: 'l_hnk', label: 'Intel NUC Hades Canyon NUC8i7HNK',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=NUC8i7HNK&_ex_kw=parts+repair+broken+motherboard+board&LH_BIN=1&LH_ItemCondition=3000&_udlo=75&_udhi=275&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=NUC8i7HNK&_ex_kw=parts+repair+broken+motherboard+board&LH_ItemCondition=3000&_udlo=75&_udhi=275&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Vega M GL \u00b7 4GB HBM2 dedicated', color: '#bb86fc' },
          { text: 'DDR4 SO-DIMM', color: '#64b5f6' },
          { text: '$250\u2013360 all-in', color: '#6be5a0' },
          { text: '\u2705 cheapest dGPU option', color: '#6be5a0' }
        ],
        note: 'Budget Hades Canyon. 15\u201320% weaker Vega M GL. Lowest entry for dedicated GPU.'
      },
      {
        id: 'l_5800h', label: 'Generic Ryzen 7 5800H Mini PC',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Ryzen+7+5800H+mini+PC&_ex_kw=laptop+notebook+parts+repair+broken+barebones&LH_BIN=1&LH_ItemCondition=3000&_udlo=50&_udhi=200&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Ryzen+7+5800H+mini+PC&_ex_kw=laptop+notebook+parts+repair+broken+barebones&LH_ItemCondition=3000&_udlo=50&_udhi=200&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Vega 8 \u00b7 GT 1030 class', color: '#bb86fc' },
          { text: 'DDR4 SO-DIMM', color: '#64b5f6' },
          { text: '$100\u2013200 w/ RAM', color: '#6be5a0' },
          { text: '\u2705 massive budget surplus', color: '#6be5a0' }
        ],
        note: 'Emulation beast (PS2/GC/Wii). Esports-only at 1080p. Cloud gaming fallback.'
      },
      {
        id: 'l_ser5pro', label: 'Beelink SER5 Pro \u2014 Ryzen 7 5800H',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER5+Pro+5800H&_ex_kw=parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=75&_udhi=250&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Beelink+SER5+Pro+5800H&_ex_kw=parts+repair+broken&LH_ItemCondition=3000&_udlo=75&_udhi=250&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Vega 8 \u00b7 GT 1030 class', color: '#bb86fc' },
          { text: 'DDR4 SO-DIMM', color: '#64b5f6' },
          { text: '$150\u2013250 w/ RAM', color: '#6be5a0' }
        ],
        note: 'Brand-name 5800H. Better thermals and resale than generics.'
      }
    ]
  },
  {
    id: 'g_emustick',
    title: 'Android Emulation Device Upgrades',
    icon: '\u{1F3AE}',
    collapsed: false,
    links: [
      {
        id: 'l_opi5', label: 'Orange Pi 5 \u2014 RK3588S',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Orange+Pi+5+RK3588&_ex_kw=parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=40&_udhi=150&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Orange+Pi+5+RK3588&_ex_kw=parts+repair+broken&LH_ItemCondition=3000&_udlo=40&_udhi=150&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Mali-G610 \u00b7 4\u00d7A76+4\u00d7A55', color: '#bb86fc' },
          { text: '4\u201316GB RAM', color: '#64b5f6' },
          { text: '$40\u2013150 used', color: '#6be5a0' },
          { text: '\u2705 best value pick', color: '#6be5a0' }
        ],
        note: 'Emu 9/10. Full PS2, GameCube, Wii, 3DS. SBC \u2014 runs Android + Linux. Native root.'
      },
      {
        id: 'l_h96rk', label: 'H96 MAX V58 \u2014 RK3588',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=H96+MAX+RK3588+Android+TV+Box&_ex_kw=parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=50&_udhi=150&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=H96+MAX+RK3588+Android+TV+Box&_ex_kw=parts+repair+broken&LH_ItemCondition=3000&_udlo=50&_udhi=150&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Mali-G610 \u00b7 4\u00d7A76+4\u00d7A55', color: '#bb86fc' },
          { text: '4\u20138GB RAM', color: '#64b5f6' },
          { text: '$50\u2013150 used', color: '#6be5a0' }
        ],
        note: 'Emu 9/10. Same RK3588 as Orange Pi 5. Turnkey TV box form factor. Easy root + TWRP.'
      },
      {
        id: 'l_shield', label: 'NVIDIA Shield TV Pro',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=NVIDIA+Shield+TV+Pro&_ex_kw=parts+repair+broken+remote+only+controller&LH_BIN=1&LH_ItemCondition=3000&_udlo=80&_udhi=200&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=NVIDIA+Shield+TV+Pro&_ex_kw=parts+repair+broken+remote+only+controller&LH_ItemCondition=3000&_udlo=80&_udhi=200&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Tegra X1+ \u00b7 256 CUDA', color: '#bb86fc' },
          { text: '3GB RAM', color: '#64b5f6' },
          { text: '$80\u2013200 used', color: '#6be5a0' },
          { text: '\u26a0 aging silicon (2019)', color: '#ff8a65' }
        ],
        note: 'Emu 8/10. Gold standard Android TV. GameCube, Wii, some PS2. 4\u00d7USB 3.0. Huge community.'
      },
      {
        id: 'l_am8', label: 'Ugoos AM8 Pro \u2014 Amlogic S928X-J',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Ugoos+AM8+Amlogic+S928X&_ex_kw=parts+repair+broken&LH_BIN=1&LH_ItemCondition=3000&_udlo=60&_udhi=180&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Ugoos+AM8+Amlogic+S928X&_ex_kw=parts+repair+broken&LH_ItemCondition=3000&_udlo=60&_udhi=180&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Mali-G57 \u00b7 A76+A55', color: '#bb86fc' },
          { text: '4GB RAM', color: '#64b5f6' },
          { text: '$60\u2013180 used', color: '#6be5a0' }
        ],
        note: 'Emu 7/10. Magisk built-in, SAMBA, NFS. Enthusiast favorite. Dreamcast, Wii, some PS2.'
      },
      {
        id: 'l_onn4k', label: 'Onn Google TV 4K Pro',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Onn+Google+TV+4K+Pro+Streaming&_ex_kw=parts+repair+broken+remote+only&LH_BIN=1&LH_ItemCondition=3000&_udlo=20&_udhi=45&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Onn+Google+TV+4K+Pro+Streaming&_ex_kw=parts+repair+broken+remote+only&LH_ItemCondition=3000&_udlo=20&_udhi=45&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'Mali-G31 \u00b7 S905X4', color: '#bb86fc' },
          { text: '3GB RAM', color: '#64b5f6' },
          { text: '$20\u201345 used', color: '#6be5a0' },
          { text: '\u26a0 marginal upgrade over X2', color: '#ff8a65' }
        ],
        note: 'Emu 6/10. Budget king. ADB out of box. PS1, N64, Dreamcast. Google TV stock Android.'
      },
      {
        id: 'l_ftv4k', label: 'Fire TV Stick 4K Max (2023)',
        urls: [
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Amazon+Fire+TV+Stick+4K+Max+2023&_ex_kw=parts+repair+broken+remote+only&LH_BIN=1&LH_ItemCondition=3000&_udlo=20&_udhi=50&LH_PrefLoc=1&_sop=15', text: 'Buy It Now' },
          { href: 'https://www.ebay.com/sch/i.html?_nkw=Amazon+Fire+TV+Stick+4K+Max+2023&_ex_kw=parts+repair+broken+remote+only&LH_ItemCondition=3000&_udlo=20&_udhi=50&LH_PrefLoc=1&LH_Sold=1&LH_Complete=1&_sop=15', text: 'Sold prices' }
        ],
        tags: [
          { text: 'GE9215 GPU \u00b7 MT8696T', color: '#bb86fc' },
          { text: '2GB RAM', color: '#64b5f6' },
          { text: '$20\u201350 used', color: '#6be5a0' },
          { text: '\u26a0 root is hard', color: '#ff8a65' }
        ],
        note: 'Emu 6/10. True stick form factor. PS1, N64, some Dreamcast. ADB sideloading only.'
      }
    ]
  }]
};

// ── Theme presets ──

var THEMES = {
  snazzy: {
    label: 'Snazzy',
    vars: {
      '--bg': '#0f0f0f', '--surface': '#1a1a1a', '--border': '#2a2a2a',
      '--text': '#e0e0e0', '--dim': '#888', '--accent': '#6be5a0',
      '--accent-dim': '#3a7a5a', '--link': '#82b1ff', '--surface-hover': '#222',
      '--tag-gpu': '#bb86fc', '--tag-ram': '#64b5f6', '--tag-price': '#6be5a0',
      '--tag-warn': '#ff8a65'
    }
  },
  aphelion: {
    label: 'Aphelion',
    vars: {
      '--bg': '#101319', '--surface': '#181c25', '--border': '#2a2e3a',
      '--text': '#f4f3ee', '--dim': '#8890a0', '--accent': '#69bfce',
      '--accent-dim': '#3a6e78', '--link': '#5679E3', '--surface-hover': '#222838',
      '--tag-gpu': '#956dca', '--tag-ram': '#69bfce', '--tag-price': '#69bfce',
      '--tag-warn': '#e37e4f'
    }
  },
  lovelace: {
    label: 'Lovelace',
    vars: {
      '--bg': '#1D1F28', '--surface': '#262833', '--border': '#363848',
      '--text': '#eaeaea', '--dim': '#8888a0', '--accent': '#C574DD',
      '--accent-dim': '#7a4088', '--link': '#8897F4', '--surface-hover': '#2e3040',
      '--tag-gpu': '#C574DD', '--tag-ram': '#79E6F3', '--tag-price': '#5ADECD',
      '--tag-warn': '#F2A272'
    }
  },
  nord: {
    label: 'Nord',
    vars: {
      '--bg': '#2E3440', '--surface': '#3B4252', '--border': '#434C5E',
      '--text': '#ECEFF4', '--dim': '#8890a0', '--accent': '#88C0D0',
      '--accent-dim': '#5E81AC', '--link': '#81A1C1', '--surface-hover': '#434C5E',
      '--tag-gpu': '#B48EAD', '--tag-ram': '#88C0D0', '--tag-price': '#A3BE8C',
      '--tag-warn': '#D08770'
    }
  }
};

// ── Bang shortcuts ──

var BANGS = {
  '!g':  'https://www.google.com/search?q=',
  '!yt': 'https://www.youtube.com/results?search_query=',
  '!gh': 'https://github.com/search?q=',
  '!r':  'https://www.reddit.com/search/?q='
};

function tryBang(value) {
  var trimmed = value.trim();
  var spaceIdx = trimmed.indexOf(' ');
  if (spaceIdx === -1) return false;
  var prefix = trimmed.slice(0, spaceIdx).toLowerCase();
  var query = trimmed.slice(spaceIdx + 1).trim();
  if (!query || !BANGS[prefix]) return false;
  window.open(BANGS[prefix] + encodeURIComponent(query), '_blank');
  return true;
}

// ── Helpers ──

function uid(prefix) {
  return prefix + '_' + Math.random().toString(36).slice(2, 9);
}

function esc(str) {
  var d = document.createElement('div');
  d.textContent = str || '';
  return d.innerHTML;
}

function dimColor(hex) {
  if (!hex || hex.length !== 7) return '#333';
  var r = parseInt(hex.slice(1, 3), 16);
  var g = parseInt(hex.slice(3, 5), 16);
  var b = parseInt(hex.slice(5, 7), 16);
  return 'rgb(' + Math.round(r * 0.4) + ',' + Math.round(g * 0.4) + ',' + Math.round(b * 0.4) + ')';
}

function debounce(fn, ms) {
  var t;
  return function() {
    var args = arguments;
    clearTimeout(t);
    t = setTimeout(function() { fn.apply(null, args); }, ms);
  };
}

function applyTheme(name) {
  var theme = THEMES[name] || THEMES.snazzy;
  var root = document.documentElement;
  var keys = Object.keys(theme.vars);
  for (var i = 0; i < keys.length; i++) {
    root.style.setProperty(keys[i], theme.vars[keys[i]]);
  }
}

// ── State management ──

var LS_KEY = 'allthelinks_data';
var DATA;
var storageAvailable = false;

function getDefaults() {
  return JSON.parse(JSON.stringify(DEFAULT_DATA));
}

function loadData() {
  try {
    var raw = localStorage.getItem(LS_KEY);
    if (!raw) return getDefaults();
    var parsed = JSON.parse(raw);
    if (!parsed.settings || !Array.isArray(parsed.groups)) return getDefaults();
    if (!parsed.settings.theme) parsed.settings.theme = 'snazzy';
    return parsed;
  } catch (e) {
    return getDefaults();
  }
}

function saveData() {
  try {
    localStorage.setItem(LS_KEY, JSON.stringify(DATA));
  } catch (e) {
    if (e.name === 'QuotaExceededError') {
      alert('Storage full. Export your data (Settings > Copy JSON) and remove some groups.');
    }
  }
}

function testStorage() {
  try {
    localStorage.setItem('__atl_test', '1');
    localStorage.removeItem('__atl_test');
    storageAvailable = true;
  } catch (e) {
    storageAvailable = false;
  }
}

// ── Rendering ──

function renderAll() {
  renderHeader();
  renderSearch();
  renderGroups();
}

function renderHeader() {
  var h = document.getElementById('header');
  h.innerHTML =
    '<div class="header-left">' +
      '<h1 class="site-title">' + esc(DATA.settings.title) + '</h1>' +
      (DATA.settings.showClock ? '<span id="clock" class="clock"></span>' : '') +
    '</div>' +
    '<div class="header-right">' +
      '<button class="btn-icon" id="btn-add-group" title="Add group">+</button>' +
      '<button class="btn-icon" id="btn-settings" title="Settings">&#9881;</button>' +
    '</div>';
  document.getElementById('btn-add-group').onclick = function() { openGroupModal(null); };
  document.getElementById('btn-settings').onclick = openSettingsModal;
  if (DATA.settings.showClock) updateClock();
}

function renderSearch() {
  var wrap = document.getElementById('search-wrap');
  if (!DATA.settings.showSearch) { wrap.innerHTML = ''; return; }
  wrap.innerHTML = '<input type="text" id="search" class="search-input" placeholder="Search links... ( / )" autocomplete="off" />';
  document.getElementById('search').addEventListener('input', debounce(onSearch, 150));
  document.getElementById('search').addEventListener('keydown', function(e) {
    if (e.key !== 'Enter') return;
    if (tryBang(this.value)) {
      this.value = '';
      onSearch();
      return;
    }
    var visibleLinks = document.querySelectorAll('.link-card:not(.hidden)');
    if (visibleLinks.length === 1) {
      var firstAnchor = visibleLinks[0].querySelector('.link-urls a');
      if (firstAnchor) window.open(firstAnchor.href, '_blank');
    }
  });
}

function renderGroups() {
  var container = document.getElementById('groups');
  container.className = DATA.settings.columns === 2 ? 'cols-2' : '';
  if (DATA.groups.length === 0) {
    container.innerHTML = '<p class="empty-state">No groups yet. Click + in the header to create one.</p>';
    return;
  }
  container.innerHTML = DATA.groups.map(renderGroup).join('');
}

function renderGroup(group) {
  var collapsed = group.collapsed ? ' collapsed' : '';
  var arrow = group.collapsed ? '\u25b8' : '\u25be';
  var linksHtml = group.links.length === 0
    ? '<p class="empty-state">No links yet. Click + to add one.</p>'
    : group.links.map(renderLink).join('');

  return '<section class="group' + collapsed + '" data-group-id="' + group.id + '">' +
    '<div class="group-header">' +
      '<button class="group-collapse" aria-label="Toggle">' + arrow + '</button>' +
      '<span class="group-icon">' + group.icon + '</span>' +
      '<h2 class="group-title">' + esc(group.title) + ' <span class="group-count">(' + group.links.length + ')</span></h2>' +
      '<div class="group-actions">' +
        '<button class="btn-icon btn-add-link" title="Add link">+</button>' +
        '<button class="btn-icon btn-edit-group" title="Edit group">&#9998;</button>' +
        '<button class="btn-icon btn-delete-group btn-danger" title="Delete group">&#128465;</button>' +
      '</div>' +
    '</div>' +
    '<div class="group-body">' + linksHtml + '</div>' +
  '</section>';
}

function renderLink(link) {
  var urlsHtml = '';
  if (link.urls && link.urls.length) {
    urlsHtml = '<div class="link-urls">' + link.urls.map(function(u) {
      var cls = (u.text && u.text.toLowerCase().indexOf('sold') !== -1) ? ' class="sold"' : '';
      return '<a href="' + esc(u.href) + '" target="_blank" rel="noopener"' + cls + '>' + esc(u.text) + '</a>';
    }).join('') + '</div>';
  }

  var tagsHtml = '';
  if (link.tags && link.tags.length) {
    tagsHtml = '<div class="link-tags">' + link.tags.map(function(t) {
      return '<span class="tag" style="color:' + t.color + ';border-color:' + dimColor(t.color) + '">' + esc(t.text) + '</span>';
    }).join('') + '</div>';
  }

  var noteHtml = link.note ? '<p class="link-note">' + esc(link.note) + '</p>' : '';

  return '<div class="link-card" data-link-id="' + link.id + '">' +
    '<div class="link-header">' +
      '<span class="link-label">' + esc(link.label) + '</span>' +
      '<div class="link-actions">' +
        '<button class="btn-icon btn-edit-link" title="Edit">&#9998;</button>' +
        '<button class="btn-icon btn-delete-link btn-danger" title="Delete">&#128465;</button>' +
      '</div>' +
    '</div>' +
    urlsHtml + tagsHtml + noteHtml +
  '</div>';
}

// ── Search ──

function onSearch() {
  var searchEl = document.getElementById('search');
  if (!searchEl) return;
  var q = searchEl.value.trim().toLowerCase();
  var groups = document.querySelectorAll('.group');

  var noResults = document.getElementById('no-results');
  if (noResults) noResults.remove();

  if (!q) {
    for (var i = 0; i < groups.length; i++) {
      groups[i].classList.remove('hidden');
      var cards = groups[i].querySelectorAll('.link-card');
      for (var j = 0; j < cards.length; j++) cards[j].classList.remove('hidden');
    }
    return;
  }

  var anyGroupVisible = false;
  for (var i = 0; i < groups.length; i++) {
    var cards = groups[i].querySelectorAll('.link-card');
    var anyCardVisible = false;
    for (var j = 0; j < cards.length; j++) {
      var match = cards[j].textContent.toLowerCase().indexOf(q) !== -1;
      cards[j].classList.toggle('hidden', !match);
      if (match) anyCardVisible = true;
    }
    groups[i].classList.toggle('hidden', !anyCardVisible);
    if (anyCardVisible) anyGroupVisible = true;
  }

  if (!anyGroupVisible) {
    var msg = document.createElement('p');
    msg.id = 'no-results';
    msg.className = 'empty-state';
    msg.innerHTML = 'No links match "' + esc(q) + '".';
    document.getElementById('groups').appendChild(msg);
  }
}

// ── Collapse/expand ──

function toggleCollapse(groupId) {
  var group = DATA.groups.find(function(g) { return g.id === groupId; });
  if (!group) return;
  group.collapsed = !group.collapsed;
  saveData();
  renderGroups();
}

// ── Clock ──

var clockInterval;

function updateClock() {
  var el = document.getElementById('clock');
  if (!el) return;
  var now = new Date();
  el.textContent = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}

function startClock() {
  updateClock();
  clearInterval(clockInterval);
  clockInterval = setInterval(updateClock, 30000);
}

// ── Event delegation ──

function initGroupDelegation() {
  var container = document.getElementById('groups');

  container.addEventListener('click', function(e) {
    var btn = e.target.closest('button');
    var groupEl, groupId, linkEl, linkId;

    if (btn) {
      groupEl = btn.closest('.group');
      groupId = groupEl ? groupEl.dataset.groupId : null;
      linkEl = btn.closest('.link-card');
      linkId = linkEl ? linkEl.dataset.linkId : null;

      if (btn.classList.contains('group-collapse')) {
        toggleCollapse(groupId);
      } else if (btn.classList.contains('btn-add-link')) {
        e.stopPropagation();
        openLinkModal(groupId, null);
      } else if (btn.classList.contains('btn-edit-group')) {
        e.stopPropagation();
        openGroupModal(groupId);
      } else if (btn.classList.contains('btn-delete-group')) {
        e.stopPropagation();
        deleteGroup(groupId);
      } else if (btn.classList.contains('btn-edit-link')) {
        openLinkModal(groupId, linkId);
      } else if (btn.classList.contains('btn-delete-link')) {
        deleteLink(groupId, linkId);
      }
      return;
    }

    // Clicking group header (not a button) toggles collapse
    var header = e.target.closest('.group-header');
    if (header && !e.target.closest('.group-actions')) {
      groupEl = header.closest('.group');
      if (groupEl) toggleCollapse(groupEl.dataset.groupId);
    }
  });
}

// ── Keyboard shortcuts ──

document.addEventListener('keydown', function(e) {
  if (e.key === '/' && !e.target.closest('input, textarea, select')) {
    e.preventDefault();
    var search = document.getElementById('search');
    if (search) search.focus();
  }
  if (e.key === 'Escape') {
    var overlay = document.getElementById('modal-overlay');
    if (!overlay.classList.contains('hidden')) {
      closeModal();
      return;
    }
    var search = document.getElementById('search');
    if (search && (document.activeElement === search || search.value)) {
      search.value = '';
      onSearch();
      search.blur();
    }
  }
});

// ── Modal system ──

function openModal(title, bodyHtml, footerHtml) {
  var overlay = document.getElementById('modal-overlay');
  overlay.querySelector('.modal-title').textContent = title;
  overlay.querySelector('.modal-body').innerHTML = bodyHtml;
  overlay.querySelector('.modal-footer').innerHTML = footerHtml;
  overlay.classList.remove('hidden');
  var firstInput = overlay.querySelector('input, textarea, select');
  if (firstInput) firstInput.focus();
}

function closeModal() {
  document.getElementById('modal-overlay').classList.add('hidden');
}

document.getElementById('modal-overlay').addEventListener('click', function(e) {
  if (e.target.classList.contains('modal-overlay')) closeModal();
});
document.querySelector('.modal-close').onclick = closeModal;

// ── CRUD: Groups ──

function addGroup(icon, title) {
  DATA.groups.push({ id: uid('g'), title: title, icon: icon, collapsed: false, links: [] });
  saveData();
  renderGroups();
}

function editGroup(groupId, icon, title) {
  var g = DATA.groups.find(function(g) { return g.id === groupId; });
  if (!g) return;
  g.icon = icon;
  g.title = title;
  saveData();
  renderGroups();
}

function deleteGroup(groupId) {
  if (!confirm('Delete this group and all its links?')) return;
  DATA.groups = DATA.groups.filter(function(g) { return g.id !== groupId; });
  saveData();
  renderGroups();
}

// ── CRUD: Links ──

function addLink(groupId, linkData) {
  var g = DATA.groups.find(function(g) { return g.id === groupId; });
  if (!g) return;
  linkData.id = uid('l');
  g.links.push(linkData);
  saveData();
  renderGroups();
}

function editLink(groupId, linkId, linkData) {
  var g = DATA.groups.find(function(g) { return g.id === groupId; });
  if (!g) return;
  var idx = g.links.findIndex(function(l) { return l.id === linkId; });
  if (idx === -1) return;
  linkData.id = linkId;
  g.links[idx] = linkData;
  saveData();
  renderGroups();
}

function deleteLink(groupId, linkId) {
  if (!confirm('Delete this link?')) return;
  var g = DATA.groups.find(function(g) { return g.id === groupId; });
  if (!g) return;
  g.links = g.links.filter(function(l) { return l.id !== linkId; });
  saveData();
  renderGroups();
}

// ── Tag parsing ──

function parseTags(str) {
  if (!str || !str.trim()) return [];
  return str.split(',').map(function(s) {
    s = s.trim();
    var match = s.match(/^(#[0-9a-fA-F]{6}):(.+)$/);
    if (match) return { color: match[1], text: match[2].trim() };
    return { color: '#6be5a0', text: s };
  }).filter(function(t) { return t.text; });
}

function serializeTags(tags) {
  if (!tags || !tags.length) return '';
  return tags.map(function(t) { return t.color + ':' + t.text; }).join(', ');
}

// ── Modal: Settings ──

function openSettingsModal() {
  var body =
    '<label>Title<input type="text" id="set-title" value="' + esc(DATA.settings.title) + '" /></label>' +
    '<label>Columns<select id="set-columns">' +
      '<option value="1"' + (DATA.settings.columns === 1 ? ' selected' : '') + '>1</option>' +
      '<option value="2"' + (DATA.settings.columns === 2 ? ' selected' : '') + '>2</option>' +
    '</select></label>' +
    '<label>Theme<select id="set-theme">' +
      Object.keys(THEMES).map(function(key) {
        var sel = DATA.settings.theme === key ? ' selected' : '';
        return '<option value="' + key + '"' + sel + '>' + THEMES[key].label + '</option>';
      }).join('') +
    '</select></label>' +
    '<label class="inline"><input type="checkbox" id="set-clock"' + (DATA.settings.showClock ? ' checked' : '') + ' /> Show clock</label>' +
    '<label class="inline"><input type="checkbox" id="set-search"' + (DATA.settings.showSearch ? ' checked' : '') + ' /> Show search</label>' +
    '<hr />' +
    '<label>JSON data (advanced)<textarea id="set-json" rows="12" spellcheck="false">' + esc(JSON.stringify(DATA, null, 2)) + '</textarea></label>';

  var footer =
    '<button class="btn" id="btn-export">Copy JSON</button>' +
    '<button class="btn btn-primary" id="btn-save-settings">Save</button>';

  openModal('Settings', body, footer);
  document.getElementById('btn-export').onclick = exportJSON;
  document.getElementById('btn-save-settings').onclick = saveSettings;
  document.getElementById('set-theme').onchange = function() {
    applyTheme(this.value);
  };
}

function saveSettings() {
  var jsonEl = document.getElementById('set-json');
  var raw = jsonEl.value;
  try {
    var parsed = JSON.parse(raw);
    if (!parsed.settings || !Array.isArray(parsed.groups)) {
      throw new Error('Invalid schema: needs settings object and groups array');
    }
    DATA = parsed;
  } catch (e) {
    alert('Invalid JSON: ' + e.message);
    return;
  }
  DATA.settings.title = document.getElementById('set-title').value.trim() || 'allthelinks';
  DATA.settings.columns = parseInt(document.getElementById('set-columns').value, 10);
  DATA.settings.showClock = document.getElementById('set-clock').checked;
  DATA.settings.showSearch = document.getElementById('set-search').checked;
  DATA.settings.theme = document.getElementById('set-theme').value || 'snazzy';
  saveData();
  closeModal();
  document.title = DATA.settings.title;
  applyTheme(DATA.settings.theme);
  renderAll();
  if (DATA.settings.showClock) startClock();
}

function exportJSON() {
  var json = JSON.stringify(DATA, null, 2);
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(json).then(
      function() { alert('Copied to clipboard.'); },
      function() { fallbackCopyAlert(); }
    );
  } else {
    fallbackCopyAlert();
  }
}

function fallbackCopyAlert() {
  var ta = document.getElementById('set-json');
  if (ta) ta.select();
  alert('Clipboard API unavailable. JSON is selected in the textarea \u2014 copy manually (Ctrl+C).');
}

// ── Modal: Group ──

function openGroupModal(groupId) {
  var existing = groupId ? DATA.groups.find(function(g) { return g.id === groupId; }) : null;
  var title = existing ? 'Edit Group' : 'Add Group';

  var body =
    '<label>Icon (emoji)<input type="text" id="grp-icon" maxlength="4" value="' + (existing ? existing.icon : '\ud83d\udcc1') + '" /></label>' +
    '<label>Title<input type="text" id="grp-title" value="' + esc(existing ? existing.title : '') + '" /></label>';

  var footer = '<button class="btn btn-primary" id="btn-save-group">Save</button>';

  openModal(title, body, footer);
  document.getElementById('btn-save-group').onclick = function() {
    var icon = document.getElementById('grp-icon').value.trim() || '\ud83d\udcc1';
    var t = document.getElementById('grp-title').value.trim();
    if (!t) { alert('Title is required.'); return; }
    if (existing) {
      editGroup(groupId, icon, t);
    } else {
      addGroup(icon, t);
    }
    closeModal();
  };
}

// ── Modal: Link ──

function openLinkModal(groupId, linkId) {
  var existing = null;
  if (linkId) {
    var g = DATA.groups.find(function(g) { return g.id === groupId; });
    if (g) existing = g.links.find(function(l) { return l.id === linkId; });
  }
  var title = existing ? 'Edit Link' : 'Add Link';

  var urlRows = '';
  if (existing && existing.urls && existing.urls.length) {
    existing.urls.forEach(function(u, i) {
      urlRows += buildUrlRow(u.text, u.href);
    });
  } else {
    urlRows = buildUrlRow('', '');
  }

  var body =
    '<label>Label<input type="text" id="lnk-label" value="' + esc(existing ? existing.label : '') + '" /></label>' +
    '<div style="font-size:0.8rem;color:var(--dim)">URLs</div>' +
    '<div id="lnk-urls-list">' + urlRows + '</div>' +
    '<button class="btn" id="btn-add-url" type="button" style="align-self:flex-start;font-size:0.75rem">+ Add URL</button>' +
    '<label>Tags <span style="font-size:0.7rem;color:var(--dim)">(comma-separated, #color:text)</span>' +
      '<input type="text" id="lnk-tags" value="' + esc(existing ? serializeTags(existing.tags) : '') + '" placeholder="#bb86fc:GPU tag, #64b5f6:RAM tag" />' +
    '</label>' +
    '<label>Note<input type="text" id="lnk-note" value="' + esc(existing ? existing.note : '') + '" /></label>';

  var footer = '<button class="btn btn-primary" id="btn-save-link">Save</button>';

  openModal(title, body, footer);

  document.getElementById('btn-add-url').onclick = function() {
    var list = document.getElementById('lnk-urls-list');
    var div = document.createElement('div');
    div.innerHTML = buildUrlRow('', '');
    list.appendChild(div.firstElementChild);
  };

  // Wire remove buttons for URL rows
  document.getElementById('lnk-urls-list').addEventListener('click', function(e) {
    var btn = e.target.closest('.btn-remove-url');
    if (!btn) return;
    var rows = document.querySelectorAll('.url-row');
    if (rows.length <= 1) return;
    btn.closest('.url-row').remove();
  });

  document.getElementById('btn-save-link').onclick = function() {
    var label = document.getElementById('lnk-label').value.trim();
    if (!label) { alert('Label is required.'); return; }

    var urls = [];
    var rows = document.querySelectorAll('.url-row');
    for (var i = 0; i < rows.length; i++) {
      var inputs = rows[i].querySelectorAll('input');
      var text = inputs[0].value.trim();
      var href = inputs[1].value.trim();
      if (href) urls.push({ text: text || 'Link', href: href });
    }

    var tags = parseTags(document.getElementById('lnk-tags').value);
    var note = document.getElementById('lnk-note').value.trim();
    var linkData = { label: label, urls: urls, tags: tags, note: note };

    if (existing) {
      editLink(groupId, linkId, linkData);
    } else {
      addLink(groupId, linkData);
    }
    closeModal();
  };
}

function buildUrlRow(text, href) {
  return '<div class="url-row">' +
    '<input type="text" placeholder="Text" value="' + esc(text) + '" style="max-width:120px" />' +
    '<input type="url" placeholder="https://..." value="' + esc(href) + '" />' +
    '<button type="button" class="btn-icon btn-remove-url btn-danger" title="Remove">&times;</button>' +
  '</div>';
}

// ── Init ──

(function init() {
  testStorage();
  DATA = loadData();
  document.title = DATA.settings.title;
  applyTheme(DATA.settings.theme || 'snazzy');

  if (!storageAvailable) {
    var warn = document.createElement('div');
    warn.className = 'warning-bar';
    warn.textContent = 'localStorage unavailable (likely Firefox file:// mode). Changes will not persist across reloads. Use Chrome or install as a browser extension for persistence.';
    document.body.insertBefore(warn, document.body.firstChild);
  }

  renderAll();
  initGroupDelegation();
  if (DATA.settings.showClock) startClock();
})();
