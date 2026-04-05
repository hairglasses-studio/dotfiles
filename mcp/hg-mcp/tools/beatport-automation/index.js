#!/usr/bin/env node
const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

const BEATPORT_URL = 'https://www.beatport.com';
const STATE_FILE = path.join(__dirname, '.beatport-state.json');
const PROGRESS_FILE = path.join(__dirname, '.beatport-progress.json');

// Progress tracker with resume capability
class ProgressTracker {
  constructor(name) {
    this.name = name;
    this.data = this.load();
  }

  load() {
    try {
      if (fs.existsSync(PROGRESS_FILE)) {
        const all = JSON.parse(fs.readFileSync(PROGRESS_FILE, 'utf8'));
        return all[this.name] || { liked: [], followed: [], failed: {} };
      }
    } catch (e) {}
    return { liked: [], followed: [], failed: {} };
  }

  save() {
    let all = {};
    try {
      if (fs.existsSync(PROGRESS_FILE)) {
        all = JSON.parse(fs.readFileSync(PROGRESS_FILE, 'utf8'));
      }
    } catch (e) {}
    all[this.name] = this.data;
    fs.writeFileSync(PROGRESS_FILE, JSON.stringify(all, null, 2));
  }

  isLiked(trackId) { return this.data.liked.includes(String(trackId)); }
  isFollowed(artistId) { return this.data.followed.includes(String(artistId)); }

  markLiked(trackId) {
    this.data.liked.push(String(trackId));
    this.save();
  }

  markFollowed(artistId) {
    this.data.followed.push(String(artistId));
    this.save();
  }

  markFailed(id, error) {
    this.data.failed[String(id)] = error;
    this.save();
  }

  summary() {
    return {
      liked: this.data.liked.length,
      followed: this.data.followed.length,
      failed: Object.keys(this.data.failed).length
    };
  }
}

class BeatportAutomation {
  constructor(playlistId = 'default') {
    this.browser = null;
    this.context = null;
    this.page = null;
    this.progress = new ProgressTracker(playlistId);
    this.startTime = Date.now();
  }

  async init(headless = true) {
    this.browser = await chromium.launch({ headless });

    // Load saved state if exists
    const stateExists = fs.existsSync(STATE_FILE);
    this.context = await this.browser.newContext(
      stateExists ? { storageState: STATE_FILE } : {}
    );
    this.page = await this.context.newPage();
    this.page.setDefaultTimeout(60000);
    this.page.setDefaultNavigationTimeout(60000);
  }

  async saveState() {
    await this.context.storageState({ path: STATE_FILE });
  }

  async close() {
    await this.saveState();
    await this.browser.close();
  }

  async dismissCookieBanner() {
    // Try to dismiss cookie consent banner
    const cookieSelectors = [
      'button:has-text("Accept")',
      'button:has-text("Accept All")',
      'button:has-text("OK")',
      'button:has-text("Got it")',
      '[data-testid="cookie-accept"]',
      '.cookie-accept',
      '#onetrust-accept-btn-handler',
    ];

    for (const selector of cookieSelectors) {
      const btn = await this.page.$(selector);
      if (btn) {
        await btn.click();
        await this.page.waitForTimeout(1000);
        console.log('Dismissed cookie banner');
        return;
      }
    }
  }

  async login(username, password) {
    console.log('Logging into Beatport...');

    // Go to main page first
    await this.page.goto(BEATPORT_URL);
    await this.page.waitForLoadState('networkidle');

    // Dismiss cookie banner
    await this.dismissCookieBanner();

    // Click login link - visible in top-right corner
    console.log('Looking for Login link...');

    // Try clicking by text content
    try {
      await this.page.click('text=Login', { timeout: 5000 });
      console.log('Clicked Login link');
    } catch (e) {
      // Try alternative approaches
      const links = await this.page.$$('a');
      for (const link of links) {
        const text = await link.textContent();
        const href = await link.getAttribute('href');
        if (text?.toLowerCase().includes('login') || href?.includes('login')) {
          console.log(`Found login: "${text}" -> ${href}`);
          await link.click();
          break;
        }
      }
    }

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(3000);

    // Take screenshot to see current state
    await this.page.screenshot({ path: '/tmp/beatport-after-login-click.png' });
    console.log('Screenshot after login click: /tmp/beatport-after-login-click.png');

    // Wait for login form with multiple possible selectors
    const emailSelectors = [
      'input[name="username"]',
      'input[name="email"]',
      'input[type="email"]',
      'input[id*="email"]',
      'input[id*="username"]',
      'input[placeholder*="Email"]',
      'input[placeholder*="Username"]',
    ];

    let emailInput = null;
    for (const selector of emailSelectors) {
      emailInput = await this.page.$(selector);
      if (emailInput) {
        console.log(`Found email input: ${selector}`);
        break;
      }
    }

    if (!emailInput) {
      // Try waiting longer and take screenshot for debugging
      await this.page.waitForTimeout(3000);
      await this.page.screenshot({ path: '/tmp/beatport-login-debug.png' });
      console.log('Screenshot saved to /tmp/beatport-login-debug.png');

      // Try again with frame handling
      const frames = this.page.frames();
      for (const frame of frames) {
        for (const selector of emailSelectors) {
          emailInput = await frame.$(selector);
          if (emailInput) {
            console.log(`Found email input in frame: ${selector}`);
            this.loginFrame = frame;
            break;
          }
        }
        if (emailInput) break;
      }
    }

    if (!emailInput) {
      throw new Error('Could not find email input field. Check /tmp/beatport-login-debug.png');
    }

    const targetFrame = this.loginFrame || this.page;

    // Fill credentials
    await emailInput.fill(username);

    const passwordInput = await targetFrame.$('input[type="password"]');
    if (!passwordInput) {
      throw new Error('Could not find password input field');
    }
    await passwordInput.fill(password);

    // Submit
    const submitButton = await targetFrame.$('button[type="submit"], input[type="submit"], button:has-text("Log In"), button:has-text("Sign In")');
    if (submitButton) {
      await submitButton.click();
    } else {
      await passwordInput.press('Enter');
    }

    // Wait for redirect/dashboard
    await this.page.waitForTimeout(5000);
    await this.page.waitForLoadState('networkidle');

    await this.saveState();
    console.log('✓ Logged in successfully');
    return true;
  }

  async isLoggedIn() {
    await this.page.goto(BEATPORT_URL);
    // Check for account menu or similar indicator
    const accountMenu = await this.page.$('[data-testid="account-menu"], .account-menu, a[href*="/account"]');
    return !!accountMenu;
  }

  async ensureLoggedIn() {
    if (!(await this.isLoggedIn())) {
      const username = process.env.BEATPORT_USERNAME;
      const password = process.env.BEATPORT_PASSWORD;
      if (!username || !password) {
        throw new Error('BEATPORT_USERNAME and BEATPORT_PASSWORD required');
      }
      await this.login(username, password);
    }
  }

  async likeTrack(trackId) {
    try {
      await this.ensureLoggedIn();

      const trackUrl = `${BEATPORT_URL}/track/-/${trackId}`;
      console.log(`Liking track ${trackId}...`);
      await this.page.goto(trackUrl);

      // Wait for page load
      await this.page.waitForLoadState('networkidle');
      await this.page.waitForTimeout(2000);

      // Dismiss any cookie banners
      await this.dismissCookieBanner();

    // Try multiple selector strategies for the add to queue/library button near the track
    // Based on screenshot: play button, queue, "+", price, share icons
    const likeSelectors = [
      // Queue/add buttons near track info
      'button[aria-label*="Add"]',
      'button[aria-label*="Queue"]',
      'button[aria-label*="Library"]',
      'button[aria-label*="Cart"]',
      '[data-testid="add-to-library"]',
      '[data-testid="add-to-cart"]',
      '[data-testid="add-to-queue"]',
      // My Beatport heart in sidebar
      'a[href*="my-beatport"]',
      '[class*="my-beatport"]',
      // Generic add/plus buttons near track title
      '.track-actions button',
      '.track-info button',
      // Heart icons
      'button:has(svg[data-icon="heart"])',
      '[class*="heart"]',
      '[class*="favorite"]',
      // The "+" button visible in screenshot
      'button:has-text("+")',
    ];

    let likeButton = null;
    for (const selector of likeSelectors) {
      likeButton = await this.page.$(selector);
      if (likeButton) {
        console.log(`Found like button: ${selector}`);
        break;
      }
    }

    // Try finding by looking at all buttons
    if (!likeButton) {
      const buttons = await this.page.$$('button');
      console.log(`Scanning ${buttons.length} buttons...`);
      for (const btn of buttons) {
        const ariaLabel = await btn.getAttribute('aria-label');
        const className = await btn.getAttribute('class');
        const title = await btn.getAttribute('title');
        const innerHTML = await btn.innerHTML();
        const text = await btn.textContent();

        // Log all buttons for debugging
        if (ariaLabel || title || (text && text.trim())) {
          console.log(`  Button: aria="${ariaLabel}", title="${title}", text="${text?.trim()}", class="${className?.slice(0,50)}"`);
        }

        if (ariaLabel?.toLowerCase().includes('library') ||
            ariaLabel?.toLowerCase().includes('like') ||
            ariaLabel?.toLowerCase().includes('favorite') ||
            ariaLabel?.toLowerCase().includes('queue') ||
            ariaLabel?.toLowerCase().includes('cart') ||
            ariaLabel?.toLowerCase().includes('add') ||
            title?.toLowerCase().includes('add') ||
            title?.toLowerCase().includes('queue') ||
            className?.toLowerCase().includes('library') ||
            innerHTML.includes('heart') ||
            innerHTML.includes('Heart')) {
          console.log(`Found button via scan: aria="${ariaLabel}", title="${title}"`);
          likeButton = btn;
          break;
        }
      }
    }

    if (likeButton) {
      // Check if already liked
      const isLiked = await likeButton.getAttribute('aria-pressed') === 'true' ||
                      await likeButton.getAttribute('data-liked') === 'true' ||
                      (await likeButton.getAttribute('class'))?.includes('active');

      if (!isLiked) {
        await likeButton.scrollIntoViewIfNeeded();
        await likeButton.click({ timeout: 10000, force: true });
        await this.page.waitForTimeout(1500);
        console.log(`✓ Liked track ${trackId}`);
        return { success: true, action: 'liked' };
      } else {
        console.log(`○ Track ${trackId} already liked`);
        return { success: true, action: 'already_liked' };
      }
    }

    console.log(`✗ Could not find like button for track ${trackId}`);
    return { success: false, error: 'like_button_not_found' };
    } catch (error) {
      console.log(`✗ Error liking track ${trackId}: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  async followArtist(artistId) {
    try {
      await this.ensureLoggedIn();

    const artistUrl = `${BEATPORT_URL}/artist/-/${artistId}`;
    console.log(`Following artist ${artistId}...`);
    await this.page.goto(artistUrl);

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(2000);

    // Dismiss any cookie banners
    await this.dismissCookieBanner();

    // Find follow button
    const followSelectors = [
      'button[aria-label*="Follow"]',
      'button[aria-label*="follow"]',
      '[data-testid="follow-button"]',
      '.follow-button',
      'button:has-text("Follow")',
      'button:has-text("follow")',
    ];

    let followButton = null;
    for (const selector of followSelectors) {
      followButton = await this.page.$(selector);
      if (followButton) {
        console.log(`Found follow button: ${selector}`);
        break;
      }
    }

    // Scan all buttons if not found
    if (!followButton) {
      const buttons = await this.page.$$('button');
      for (const btn of buttons) {
        const text = await btn.textContent();
        const ariaLabel = await btn.getAttribute('aria-label');
        if (text?.toLowerCase().includes('follow') || ariaLabel?.toLowerCase().includes('follow')) {
          console.log(`Found follow button via scan: "${text}"`);
          followButton = btn;
          break;
        }
      }
    }

    if (followButton) {
      const buttonText = await followButton.textContent();
      const isFollowing = buttonText?.toLowerCase().includes('following') ||
                          await followButton.getAttribute('data-following') === 'true' ||
                          (await followButton.getAttribute('class'))?.includes('active');

      if (!isFollowing) {
        await followButton.scrollIntoViewIfNeeded();
        await followButton.click({ timeout: 10000, force: true });
        await this.page.waitForTimeout(1500);
        console.log(`✓ Followed artist ${artistId}`);
        return { success: true, action: 'followed' };
      } else {
        console.log(`○ Already following artist ${artistId}`);
        return { success: true, action: 'already_following' };
      }
    }

    console.log(`✗ Could not find follow button for artist ${artistId}`);
    return { success: false, error: 'follow_button_not_found' };
    } catch (error) {
      console.log(`✗ Error following artist ${artistId}: ${error.message}`);
      return { success: false, error: error.message };
    }
  }

  async likeTracksFromPlaylist(playlistId) {
    await this.ensureLoggedIn();

    const playlistUrl = `${BEATPORT_URL}/playlist/-/${playlistId}`;
    console.log(`Processing playlist ${playlistId}...`);
    await this.page.goto(playlistUrl);

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(3000);

    // Scroll to load all tracks (lazy loading)
    console.log('Scrolling to load all tracks...');
    let previousHeight = 0;
    for (let i = 0; i < 50; i++) {
      await this.page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      await this.page.waitForTimeout(1000);
      const currentHeight = await this.page.evaluate(() => document.body.scrollHeight);
      if (currentHeight === previousHeight) break;
      previousHeight = currentHeight;
    }
    await this.page.evaluate(() => window.scrollTo(0, 0));

    // Take debug screenshot
    await this.page.screenshot({ path: `/tmp/beatport-playlist-${playlistId}.png` });

    // Get all track links with multiple selector strategies
    let trackLinks = await this.page.$$eval(
      'a[href*="/track/"]',
      links => links.map(a => {
        const match = a.href.match(/\/track\/[^/]+\/(\d+)/);
        return match ? match[1] : null;
      }).filter(Boolean)
    );

    // Try alternative: look for track IDs in data attributes
    if (trackLinks.length === 0) {
      trackLinks = await this.page.$$eval(
        '[data-track-id], [data-id]',
        els => els.map(el => el.getAttribute('data-track-id') || el.getAttribute('data-id')).filter(Boolean)
      );
    }

    const uniqueTracks = [...new Set(trackLinks)];
    console.log(`Found ${uniqueTracks.length} tracks in playlist`);

    if (uniqueTracks.length === 0) {
      console.log(`Debug screenshot: /tmp/beatport-playlist-${playlistId}.png`);
    }

    const results = { liked: 0, already_liked: 0, failed: 0 };

    for (const trackId of uniqueTracks) {
      const result = await this.likeTrack(trackId);
      if (result.action === 'liked') results.liked++;
      else if (result.action === 'already_liked') results.already_liked++;
      else results.failed++;

      // Rate limit
      await this.page.waitForTimeout(2000);
    }

    return results;
  }

  async followArtistsFromPlaylist(playlistId) {
    await this.ensureLoggedIn();

    const playlistUrl = `${BEATPORT_URL}/playlist/-/${playlistId}`;
    console.log(`Getting artists from playlist ${playlistId}...`);
    await this.page.goto(playlistUrl);

    await this.page.waitForLoadState('networkidle');
    await this.page.waitForTimeout(3000);

    // Scroll to load all content
    console.log('Scrolling to load all artists...');
    let previousHeight = 0;
    for (let i = 0; i < 50; i++) {
      await this.page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      await this.page.waitForTimeout(1000);
      const currentHeight = await this.page.evaluate(() => document.body.scrollHeight);
      if (currentHeight === previousHeight) break;
      previousHeight = currentHeight;
    }
    await this.page.evaluate(() => window.scrollTo(0, 0));

    // Get all artist links
    const artistLinks = await this.page.$$eval(
      'a[href*="/artist/"]',
      links => links.map(a => {
        const match = a.href.match(/\/artist\/[^/]+\/(\d+)/);
        return match ? match[1] : null;
      }).filter(Boolean)
    );

    const uniqueArtists = [...new Set(artistLinks)];
    console.log(`Found ${uniqueArtists.length} artists in playlist`);

    if (uniqueArtists.length === 0) {
      console.log(`Debug screenshot: /tmp/beatport-playlist-${playlistId}.png`);
    }

    const results = { followed: 0, already_following: 0, failed: 0 };

    for (const artistId of uniqueArtists) {
      const result = await this.followArtist(artistId);
      if (result.action === 'followed') results.followed++;
      else if (result.action === 'already_following') results.already_following++;
      else results.failed++;

      // Rate limit
      await this.page.waitForTimeout(2000);
    }

    return results;
  }
}

// Format elapsed time
function formatElapsed(ms) {
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  const h = Math.floor(m / 60);
  if (h > 0) return `${h}h${m % 60}m`;
  if (m > 0) return `${m}m${s % 60}s`;
  return `${s}s`;
}

// CLI interface
async function main() {
  const args = process.argv.slice(2);
  const command = args[0];

  if (!command) {
    console.log(`
Beatport Automation CLI

Usage:
  node index.js login                      - Login and save session
  node index.js like-track <trackId>       - Like a single track
  node index.js like-tracks <id1,id2,...>  - Like multiple tracks (comma-separated)
  node index.js follow-artist <artistId>   - Follow a single artist
  node index.js follow-artists <id1,id2,...> - Follow multiple artists (comma-separated)
  node index.js status                     - Show progress status
  node index.js reset                      - Reset progress tracking

Options:
  --headless=false  - Show browser window
  --playlist=ID     - Track progress per playlist (for resume)
`);
    process.exit(0);
  }

  const headless = !args.includes('--headless=false');
  const playlistArg = args.find(a => a.startsWith('--playlist='));
  const playlistId = playlistArg ? playlistArg.split('=')[1] : 'default';
  const automation = new BeatportAutomation(playlistId);

  try {
    switch (command) {
      case 'status': {
        const summary = automation.progress.summary();
        console.log('Progress:', summary);
        process.exit(0);
      }

      case 'reset': {
        fs.unlinkSync(PROGRESS_FILE);
        console.log('Progress reset');
        process.exit(0);
      }
    }

    await automation.init(headless);

    switch (command) {
      case 'login':
        await automation.ensureLoggedIn();
        console.log('Session saved');
        break;

      case 'like-track':
        await automation.likeTrack(args[1]);
        break;

      case 'like-tracks': {
        const trackIds = args[1].split(',').filter(Boolean);
        const toProcess = trackIds.filter(id => !automation.progress.isLiked(id.trim()));
        const skipped = trackIds.length - toProcess.length;

        console.log(`Liking ${toProcess.length} tracks (${skipped} already done)...`);
        const results = { liked: 0, already_liked: 0, failed: 0, skipped };
        const startTime = Date.now();

        for (let i = 0; i < toProcess.length; i++) {
          const trackId = toProcess[i].trim();
          const result = await automation.likeTrack(trackId);

          if (result.action === 'liked') {
            results.liked++;
            automation.progress.markLiked(trackId);
          } else if (result.action === 'already_liked') {
            results.already_liked++;
            automation.progress.markLiked(trackId);
          } else {
            results.failed++;
            automation.progress.markFailed(trackId, result.error);
          }

          // Progress update every 10 tracks
          if ((i + 1) % 10 === 0) {
            const elapsed = Date.now() - startTime;
            const rate = (i + 1) / (elapsed / 60000);
            const eta = (toProcess.length - i - 1) / rate;
            process.stdout.write(`\r[${i + 1}/${toProcess.length}] ${rate.toFixed(1)}/min, ETA: ${formatElapsed(eta * 60000)}    `);
          }

          await automation.page.waitForTimeout(1500);
        }
        console.log('\nResults:', results);
        break;
      }

      case 'follow-artist':
        await automation.followArtist(args[1]);
        break;

      case 'follow-artists': {
        const artistIds = args[1].split(',').filter(Boolean);
        const toProcess = artistIds.filter(id => !automation.progress.isFollowed(id.trim()));
        const skipped = artistIds.length - toProcess.length;

        console.log(`Following ${toProcess.length} artists (${skipped} already done)...`);
        const results = { followed: 0, already_following: 0, failed: 0, skipped };
        const startTime = Date.now();

        for (let i = 0; i < toProcess.length; i++) {
          const artistId = toProcess[i].trim();
          const result = await automation.followArtist(artistId);

          if (result.action === 'followed') {
            results.followed++;
            automation.progress.markFollowed(artistId);
          } else if (result.action === 'already_following') {
            results.already_following++;
            automation.progress.markFollowed(artistId);
          } else {
            results.failed++;
            automation.progress.markFailed(artistId, result.error);
          }

          // Progress update every 10 artists
          if ((i + 1) % 10 === 0) {
            const elapsed = Date.now() - startTime;
            const rate = (i + 1) / (elapsed / 60000);
            const eta = (toProcess.length - i - 1) / rate;
            process.stdout.write(`\r[${i + 1}/${toProcess.length}] ${rate.toFixed(1)}/min, ETA: ${formatElapsed(eta * 60000)}    `);
          }

          await automation.page.waitForTimeout(1500);
        }
        console.log('\nResults:', results);
        break;
      }

      default:
        console.error(`Unknown command: ${command}`);
        process.exit(1);
    }
  } catch (error) {
    console.error('Error:', error.message);
    process.exit(1);
  } finally {
    await automation.close();
  }
}

main();

module.exports = { BeatportAutomation };
