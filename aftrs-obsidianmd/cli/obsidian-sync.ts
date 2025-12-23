#!/usr/bin/env node

import { promises as fs } from 'fs';
import { dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

interface ObsidianConfig {
  vaultPath: string;
  supermemoryApiKey: string;
  supermemoryApiUrl: string;
  autoSync: boolean;
  syncInterval: number; // minutes
  includeTypes: string[];
  preserveMetadata: boolean;
}

interface SupermemoryMemory {
  id: string;
  title: string;
  content: string;
  url?: string;
  createdAt: string;
  updatedAt: string;
  tags: string[];
  type: 'note' | 'link' | 'file' | 'chat';
  metadata: {
    source?: string;
    author?: string;
    description?: string;
  };
}

interface SupermemoryResponse {
  memories: SupermemoryMemory[];
  total: number;
  page: number;
  hasMore: boolean;
}

class ObsidianSync {
  private config: ObsidianConfig;

  constructor(config: ObsidianConfig) {
    this.config = config;
  }

  async sync(): Promise<void> {
    console.log('🔗 Starting Obsidian-Supermemory sync...');

    // Verify vault exists
    await this.verifyVault();

    // Test Supermemory connection
    await this.testSupermemoryConnection();

    // Create sync folders
    const syncDir = join(this.config.vaultPath, 'Supermemory');
    await fs.mkdir(syncDir, { recursive: true });

    console.log('🔄 Syncing memories from Supermemory...');

    let totalImported = 0;
    let page = 1;
    let hasMore = true;

    while (hasMore) {
      console.log(`📖 Fetching memories page ${page}...`);

      const response = await this.fetchMemories(page, 50);

      for (const memory of response.memories) {
        if (this.config.includeTypes.includes(memory.type)) {
          await this.saveMemoryToVault(memory, syncDir);
          totalImported++;
          process.stdout.write(`\r✅ Synced ${totalImported} memories`);
        }
      }

      hasMore = response.hasMore;
      page++;

      // Rate limiting
      if (hasMore) {
        await new Promise(resolve => setTimeout(resolve, 1000));
      }
    }

    console.log(`\n🎉 Sync complete! Added ${totalImported} memories to vault at ${this.config.vaultPath}`);
  }

  async startAutoSync(): Promise<void> {
    console.log(`🔄 Starting auto-sync every ${this.config.syncInterval} minutes...`);

    // Initial sync
    await this.sync();

    // Schedule periodic syncs
    setInterval(async () => {
      console.log('\n⏰ Running scheduled sync...');
      try {
        await this.sync();
      } catch (error) {
        console.error('❌ Auto-sync failed:', (error as Error).message);
      }
    }, this.config.syncInterval * 60 * 1000);

    // Keep process alive
    console.log('✨ Auto-sync is running. Press Ctrl+C to stop.');
    process.on('SIGINT', () => {
      console.log('\n👋 Stopping auto-sync...');
      process.exit(0);
    });
  }

  private async verifyVault(): Promise<void> {
    try {
      const stats = await fs.stat(this.config.vaultPath);
      if (!stats.isDirectory()) {
        throw new Error('Vault path is not a directory');
      }
    } catch (error) {
      throw new Error(`Invalid vault path: ${this.config.vaultPath}`);
    }

    console.log('✅ Obsidian vault verified');
  }

  private async testSupermemoryConnection(): Promise<void> {
    const response = await fetch(`${this.config.supermemoryApiUrl}/v1/memories?limit=1`, {
      headers: {
        'Authorization': `Bearer ${this.config.supermemoryApiKey}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`Supermemory API connection failed: ${response.status} ${response.statusText}`);
    }

    console.log('✅ Connected to Supermemory API');
  }

  private async fetchMemories(page: number = 1, limit: number = 100): Promise<SupermemoryResponse> {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      types: this.config.includeTypes.join(',')
    });

    const response = await fetch(`${this.config.supermemoryApiUrl}/v1/memories?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.config.supermemoryApiKey}`,
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch memories: ${response.status} ${response.statusText}`);
    }

    return await response.json();
  }

  private formatMemoryContent(memory: SupermemoryMemory): string {
    let content = '';

    // Add frontmatter if preserving metadata
    if (this.config.preserveMetadata) {
      content += '---\n';
      content += `supermemory_id: ${memory.id}\n`;
      content += `type: ${memory.type}\n`;
      content += `created: ${memory.createdAt}\n`;
      content += `updated: ${memory.updatedAt}\n`;

      if (memory.url) {
        content += `url: ${memory.url}\n`;
      }

      if (memory.tags.length > 0) {
        content += `tags:\n${memory.tags.map(tag => `  - ${tag}`).join('\n')}\n`;
      }

      if (memory.metadata.source) {
        content += `source: ${memory.metadata.source}\n`;
      }

      if (memory.metadata.author) {
        content += `author: ${memory.metadata.author}\n`;
      }

      content += '---\n\n';
    }

    // Add title
    content += `# ${memory.title}\n\n`;

    // Add metadata as content if not in frontmatter
    if (!this.config.preserveMetadata && memory.url) {
      content += `**Source:** ${memory.url}\n\n`;
    }

    if (!this.config.preserveMetadata && memory.tags.length > 0) {
      content += `**Tags:** ${memory.tags.map(tag => `#${tag.replace(/\s+/g, '_')}`).join(' ')}\n\n`;
    }

    // Add main content
    content += memory.content;

    // Add creation date at bottom if not in frontmatter
    if (!this.config.preserveMetadata) {
      content += `\n\n---\n*Imported from Supermemory on ${new Date().toISOString().split('T')[0]}*\n`;
      content += `*Originally created: ${memory.createdAt}*`;
    }

    return content;
  }

  private sanitizeFilename(title: string): string {
    return title
      .replace(/[:|?<>*\\]/g, '')
      .replace(/\//g, '-')
      .substring(0, 100)
      .trim();
  }

  private async saveMemoryToVault(memory: SupermemoryMemory, syncDir: string): Promise<void> {
    const content = this.formatMemoryContent(memory);
    const filename = this.sanitizeFilename(memory.title);

    // Create subfolder by type if preserving metadata
    let targetDir = syncDir;
    if (this.config.preserveMetadata) {
      targetDir = join(syncDir, `${memory.type}s`);
      await fs.mkdir(targetDir, { recursive: true });
    }

    const filepath = join(targetDir, `${filename}.md`);
    await fs.writeFile(filepath, content, 'utf-8');
  }

  async createObsidianPlugin(): Promise<void> {
    console.log('🔌 Creating Obsidian plugin configuration...');

    const pluginDir = join(this.config.vaultPath, '.obsidian', 'plugins', 'supermemory-sync');
    await fs.mkdir(pluginDir, { recursive: true });

    const manifest = {
      id: 'supermemory-sync',
      name: 'Supermemory Sync',
      version: '1.0.0',
      minAppVersion: '0.15.0',
      description: 'Automatically sync memories from Supermemory.ai',
      author: 'AFTRS',
      authorUrl: 'https://github.com/aftrs-void',
      isDesktopOnly: false
    };

    await fs.writeFile(
      join(pluginDir, 'manifest.json'),
      JSON.stringify(manifest, null, 2),
      'utf-8'
    );

    const pluginConfig = {
      supermemoryApiKey: this.config.supermemoryApiKey,
      supermemoryApiUrl: this.config.supermemoryApiUrl,
      autoSync: this.config.autoSync,
      syncInterval: this.config.syncInterval,
      includeTypes: this.config.includeTypes,
      preserveMetadata: this.config.preserveMetadata
    };

    await fs.writeFile(
      join(pluginDir, 'data.json'),
      JSON.stringify(pluginConfig, null, 2),
      'utf-8'
    );

    console.log('✅ Obsidian plugin configuration created');
  }
}

// CLI interface
async function main() {
  const args = process.argv.slice(2);

  const config: ObsidianConfig = {
    vaultPath: process.env.OBSIDIAN_VAULT_PATH || '',
    supermemoryApiKey: process.env.SUPERMEMORY_API_KEY || '',
    supermemoryApiUrl: process.env.SUPERMEMORY_API_URL || 'https://api.supermemory.ai',
    autoSync: false,
    syncInterval: 30, // minutes
    includeTypes: ['note', 'link', 'file', 'chat'],
    preserveMetadata: true
  };

  // Parse command line arguments
  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--vault-path':
        config.vaultPath = args[++i];
        break;
      case '--api-key':
        config.supermemoryApiKey = args[++i];
        break;
      case '--api-url':
        config.supermemoryApiUrl = args[++i];
        break;
      case '--types':
        config.includeTypes = args[++i].split(',');
        break;
      case '--no-metadata':
        config.preserveMetadata = false;
        break;
      case '--auto-sync':
        config.autoSync = true;
        if (args[i + 1] && !args[i + 1].startsWith('--')) {
          config.syncInterval = parseInt(args[++i]);
        }
        break;
      case '--help':
        console.log(`
Obsidian-Supermemory Sync Tool

Usage: obsidian-sync [options] [command]

Commands:
  sync         Sync memories once (default)
  auto-sync    Start continuous auto-sync
  setup-plugin Create Obsidian plugin configuration

Options:
  --vault-path <path>    Path to Obsidian vault (or set OBSIDIAN_VAULT_PATH env var)
  --api-key <key>        Supermemory API key (or set SUPERMEMORY_API_KEY env var)
  --api-url <url>        Supermemory API URL (default: https://api.supermemory.ai)
  --types <types>        Comma-separated list of types to include (default: note,link,file,chat)
  --no-metadata          Don't include metadata in frontmatter
  --auto-sync [minutes]  Enable auto-sync with optional interval (default: 30 minutes)
  --help                 Show this help message

Examples:
  obsidian-sync --vault-path ~/Documents/MyVault sync
  obsidian-sync --api-key sm_abc123 --auto-sync 60
  obsidian-sync setup-plugin
        `);
        process.exit(0);
        break;
    }
  }

  const command = args[args.length - 1];

  if (!config.vaultPath) {
    console.error('❌ Error: Vault path is required. Use --vault-path or set OBSIDIAN_VAULT_PATH environment variable');
    process.exit(1);
  }

  if (!config.supermemoryApiKey) {
    console.error('❌ Error: Supermemory API key is required. Use --api-key or set SUPERMEMORY_API_KEY environment variable');
    process.exit(1);
  }

  try {
    const sync = new ObsidianSync(config);

    switch (command) {
      case 'auto-sync':
        await sync.startAutoSync();
        break;
      case 'setup-plugin':
        await sync.createObsidianPlugin();
        break;
      case 'sync':
      default:
        await sync.sync();
        break;
    }
  } catch (error) {
    console.error('❌ Error:', (error as Error).message);
    process.exit(1);
  }
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}
