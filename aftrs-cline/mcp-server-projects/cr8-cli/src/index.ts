#!/usr/bin/env node

/**
 * MCP Server for CR8 CLI - Professional DJ Media Processing
 * Provides AI access to Beatport integration, audio analysis, and media workflows
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ErrorCode,
  ListToolsRequestSchema,
  McpError,
} from '@modelcontextprotocol/sdk/types.js';
import { spawn } from 'child_process';
import sqlite3 from 'sqlite3';

// CR8 CLI base path
const CR8_CLI_PATH = process.env.CR8_CLI_PATH || '/home/hg/Docs/aftrs-void/cr8_cli';

interface PlaylistInfo {
  service: string;
  user: string;
  url: string;
  title?: string;
  trackCount?: number;
  lastSync?: string;
}

interface TrackMetadata {
  title: string;
  artist: string;
  bpm?: number;
  key?: string;
  genre?: string;
  duration?: number;
  url?: string;
  beatportId?: string;
  confidence?: number;
}

interface DownloadResult {
  success: boolean;
  filename?: string;
  error?: string;
  metadata?: TrackMetadata;
}

class CR8MCPServer {
  private server: Server;
  private db: sqlite3.Database | null = null;

  constructor() {
    this.server = new Server(
      {
        name: 'cr8-cli',
        version: '1.0.0',
      },
      {
        capabilities: {
          tools: {},
        },
      }
    );

    this.setupHandlers();
  }

  private setupHandlers() {
    this.server.setRequestHandler(ListToolsRequestSchema, async () => {
      return {
        tools: [
          // Media Download Tools
          {
            name: 'download_track',
            description: 'Download a single track from YouTube, SoundCloud, or Spotify with metadata enhancement',
            inputSchema: {
              type: 'object',
              properties: {
                url: {
                  type: 'string',
                  description: 'URL of the track to download',
                },
                quality: {
                  type: 'string',
                  enum: ['320k', 'best', 'high'],
                  description: 'Audio quality preference',
                  default: '320k',
                },
                enhance_metadata: {
                  type: 'boolean',
                  description: 'Apply Beatport metadata enhancement',
                  default: true,
                },
              },
              required: ['url'],
            },
          },
          {
            name: 'download_playlist',
            description: 'Download entire playlist with batch processing and quality assurance',
            inputSchema: {
              type: 'object',
              properties: {
                url: {
                  type: 'string',
                  description: 'Playlist URL (YouTube, SoundCloud, Spotify)',
                },
                user: {
                  type: 'string',
                  description: 'User/artist name for organization',
                },
                sync_to_drive: {
                  type: 'boolean',
                  description: 'Sync to Google Drive after download',
                  default: true,
                },
              },
              required: ['url', 'user'],
            },
          },
          // Audio Analysis Tools
          {
            name: 'analyze_audio',
            description: 'Perform comprehensive audio analysis (BPM, key, genre)',
            inputSchema: {
              type: 'object',
              properties: {
                file_path: {
                  type: 'string',
                  description: 'Path to audio file for analysis',
                },
                analysis_type: {
                  type: 'string',
                  enum: ['bpm', 'key', 'genre', 'all'],
                  description: 'Type of analysis to perform',
                  default: 'all',
                },
              },
              required: ['file_path'],
            },
          },
          {
            name: 'beatport_lookup',
            description: 'Search Beatport database for professional DJ metadata',
            inputSchema: {
              type: 'object',
              properties: {
                query: {
                  type: 'string',
                  description: 'Search query (artist - title)',
                },
                track_file: {
                  type: 'string',
                  description: 'Audio file path for fingerprint matching',
                },
                filters: {
                  type: 'object',
                  properties: {
                    genre: { type: 'string' },
                    label: { type: 'string' },
                    bpm_range: {
                      type: 'object',
                      properties: {
                        min: { type: 'number' },
                        max: { type: 'number' },
                      },
                    },
                  },
                },
              },
              required: ['query'],
            },
          },
          // Playlist Management Tools
          {
            name: 'list_playlists',
            description: 'List all registered playlists in CR8 system',
            inputSchema: {
              type: 'object',
              properties: {
                service: {
                  type: 'string',
                  enum: ['soundcloud', 'youtube', 'spotify', 'all'],
                  description: 'Filter by service type',
                  default: 'all',
                },
                status: {
                  type: 'string',
                  enum: ['active', 'archived', 'error', 'all'],
                  description: 'Filter by status',
                  default: 'all',
                },
              },
            },
          },
          {
            name: 'sync_playlists',
            description: 'Sync all or specific playlists with latest tracks',
            inputSchema: {
              type: 'object',
              properties: {
                playlist_ids: {
                  type: 'array',
                  items: { type: 'string' },
                  description: 'Specific playlist IDs to sync (empty = all)',
                },
                force: {
                  type: 'boolean',
                  description: 'Force re-download even if already synced',
                  default: false,
                },
              },
            },
          },
          // Database & Cache Management
          {
            name: 'query_database',
            description: 'Query CR8 database for tracks, playlists, or metadata',
            inputSchema: {
              type: 'object',
              properties: {
                query_type: {
                  type: 'string',
                  enum: ['tracks', 'playlists', 'metadata', 'stats', 'custom'],
                  description: 'Type of query to execute',
                },
                filters: {
                  type: 'object',
                  properties: {
                    artist: { type: 'string' },
                    genre: { type: 'string' },
                    bpm_range: {
                      type: 'object',
                      properties: {
                        min: { type: 'number' },
                        max: { type: 'number' },
                      },
                    },
                    key: { type: 'string' },
                    date_added: { type: 'string' },
                  },
                },
                limit: {
                  type: 'number',
                  description: 'Maximum results to return',
                  default: 50,
                },
                custom_sql: {
                  type: 'string',
                  description: 'Custom SQL query (for query_type: custom)',
                },
              },
              required: ['query_type'],
            },
          },
          {
            name: 'clear_cache',
            description: 'Clear download cache, metadata cache, or both',
            inputSchema: {
              type: 'object',
              properties: {
                cache_type: {
                  type: 'string',
                  enum: ['download', 'metadata', 'beatport', 'all'],
                  description: 'Type of cache to clear',
                  default: 'all',
                },
                confirm: {
                  type: 'boolean',
                  description: 'Confirmation required for destructive operation',
                  default: false,
                },
              },
            },
          },
          // System Status & Health
          {
            name: 'system_status',
            description: 'Get CR8 CLI system status and health information',
            inputSchema: {
              type: 'object',
              properties: {
                include_stats: {
                  type: 'boolean',
                  description: 'Include detailed statistics',
                  default: true,
                },
              },
            },
          },
          {
            name: 'run_diagnostics',
            description: 'Run comprehensive diagnostics on CR8 CLI system',
            inputSchema: {
              type: 'object',
              properties: {
                check_dependencies: {
                  type: 'boolean',
                  description: 'Check system dependencies',
                  default: true,
                },
                test_apis: {
                  type: 'boolean',
                  description: 'Test API connectivity',
                  default: true,
                },
                validate_database: {
                  type: 'boolean',
                  description: 'Validate database integrity',
                  default: true,
                },
              },
            },
          },
          // Advanced Workflow Tools
          {
            name: 'enhance_collection',
            description: 'Batch enhance existing collection with professional metadata',
            inputSchema: {
              type: 'object',
              properties: {
                directory: {
                  type: 'string',
                  description: 'Directory path containing audio files',
                },
                enhancement_type: {
                  type: 'string',
                  enum: ['beatport', 'beets', 'ml', 'all'],
                  description: 'Enhancement method to apply',
                  default: 'all',
                },
                dry_run: {
                  type: 'boolean',
                  description: 'Preview changes without applying',
                  default: false,
                },
              },
              required: ['directory'],
            },
          },
          {
            name: 'export_for_rekordbox',
            description: 'Export playlist or collection for Rekordbox DJ software',
            inputSchema: {
              type: 'object',
              properties: {
                playlist_name: {
                  type: 'string',
                  description: 'Name of playlist to export',
                },
                export_format: {
                  type: 'string',
                  enum: ['xml', 'm3u8', 'crate'],
                  description: 'Export format for Rekordbox',
                  default: 'xml',
                },
                include_cue_points: {
                  type: 'boolean',
                  description: 'Include cue point data',
                  default: true,
                },
              },
              required: ['playlist_name'],
            },
          },
        ],
      };
    });

    this.server.setRequestHandler(CallToolRequestSchema, async (request) => {
      try {
        const { name, arguments: args } = request.params;

        switch (name) {
          case 'download_track':
            return await this.downloadTrack(args.url, args.quality, args.enhance_metadata);

          case 'download_playlist':
            return await this.downloadPlaylist(args.url, args.user, args.sync_to_drive);

          case 'analyze_audio':
            return await this.analyzeAudio(args.file_path, args.analysis_type);

          case 'beatport_lookup':
            return await this.beatportLookup(args.query, args.track_file, args.filters);

          case 'list_playlists':
            return await this.listPlaylists(args.service, args.status);

          case 'sync_playlists':
            return await this.syncPlaylists(args.playlist_ids, args.force);

          case 'query_database':
            return await this.queryDatabase(args.query_type, args.filters, args.limit, args.custom_sql);

          case 'clear_cache':
            return await this.clearCache(args.cache_type, args.confirm);

          case 'system_status':
            return await this.getSystemStatus(args.include_stats);

          case 'run_diagnostics':
            return await this.runDiagnostics(args.check_dependencies, args.test_apis, args.validate_database);

          case 'enhance_collection':
            return await this.enhanceCollection(args.directory, args.enhancement_type, args.dry_run);

          case 'export_for_rekordbox':
            return await this.exportForRekordbox(args.playlist_name, args.export_format, args.include_cue_points);

          default:
            throw new McpError(ErrorCode.MethodNotFound, `Unknown tool: ${name}`);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
        throw new McpError(ErrorCode.InternalError, `CR8 CLI error: ${errorMessage}`);
      }
    });
  }

  private async execCR8Command(command: string[]): Promise<{ stdout: string; stderr: string; exitCode: number }> {
    return new Promise((resolve) => {
      const cr8Process = spawn('./cr8', command, {
        cwd: CR8_CLI_PATH,
        stdio: ['pipe', 'pipe', 'pipe'],
      });

      let stdout = '';
      let stderr = '';

      cr8Process.stdout?.on('data', (data) => {
        stdout += data.toString();
      });

      cr8Process.stderr?.on('data', (data) => {
        stderr += data.toString();
      });

      cr8Process.on('close', (exitCode) => {
        resolve({ stdout, stderr, exitCode: exitCode || 0 });
      });
    });
  }

  private async downloadTrack(url: string, quality = '320k', enhanceMetadata = true): Promise<any> {
    const command = ['download', url];

    if (quality !== '320k') {
      command.push('--quality', quality);
    }

    if (enhanceMetadata) {
      command.push('--enhance');
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Download Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async downloadPlaylist(url: string, user: string, syncToDrive = true): Promise<any> {
    const command = ['gdrive-sync', user, url];

    if (!syncToDrive) {
      command.push('--local-only');
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Playlist Download Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async analyzeAudio(filePath: string, analysisType = 'all'): Promise<any> {
    const command = ['audio-intelligence', 'analyze', filePath];

    if (analysisType !== 'all') {
      command.push('--type', analysisType);
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Audio Analysis Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async beatportLookup(query: string, trackFile?: string, filters?: any): Promise<any> {
    const command = ['beatport', 'search', query];

    if (trackFile) {
      command.push('--audio-file', trackFile);
    }

    if (filters?.genre) {
      command.push('--genre', filters.genre);
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Beatport Lookup Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async listPlaylists(service = 'all', status = 'all'): Promise<any> {
    const command = ['crate', 'list'];

    if (service !== 'all') {
      command.push('--service', service);
    }

    if (status !== 'all') {
      command.push('--status', status);
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Playlists:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async syncPlaylists(playlistIds?: string[], force = false): Promise<any> {
    const command = ['sync'];

    if (playlistIds && playlistIds.length > 0) {
      command.push('--playlists', playlistIds.join(','));
    }

    if (force) {
      command.push('--force');
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Sync Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async queryDatabase(queryType: string, filters?: any, limit = 50, customSql?: string): Promise<any> {
    let command: string[];

    if (queryType === 'custom' && customSql) {
      command = ['db', 'query', '--sql', customSql];
    } else {
      command = ['db', 'query', '--type', queryType];

      if (filters) {
        if (filters.artist) command.push('--artist', filters.artist);
        if (filters.genre) command.push('--genre', filters.genre);
        if (filters.key) command.push('--key', filters.key);
      }

      command.push('--limit', limit.toString());
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Database Query Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async clearCache(cacheType = 'all', confirm = false): Promise<any> {
    if (!confirm) {
      return {
        content: [
          {
            type: 'text',
            text: 'Cache clearing requires confirmation. Set confirm=true to proceed.',
          },
        ],
      };
    }

    const command = ['cache', 'clear', '--type', cacheType, '--force'];
    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Cache Clear Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async getSystemStatus(includeStats = true): Promise<any> {
    const command = ['status'];

    if (includeStats) {
      command.push('--stats');
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `System Status:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async runDiagnostics(checkDeps = true, testApis = true, validateDb = true): Promise<any> {
    const command = ['doctor'];
    const options: string[] = [];

    if (checkDeps) options.push('--check-deps');
    if (testApis) options.push('--test-apis');
    if (validateDb) options.push('--validate-db');

    const result = await this.execCR8Command([...command, ...options]);

    return {
      content: [
        {
          type: 'text',
          text: `Diagnostics Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async enhanceCollection(directory: string, enhancementType = 'all', dryRun = false): Promise<any> {
    const command = ['batch', 'enhance', directory];

    if (enhancementType !== 'all') {
      command.push('--method', enhancementType);
    }

    if (dryRun) {
      command.push('--dry-run');
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Enhancement Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  private async exportForRekordbox(playlistName: string, exportFormat = 'xml', includeCuePoints = true): Promise<any> {
    const command = ['rekordbox', 'export', playlistName, '--format', exportFormat];

    if (includeCuePoints) {
      command.push('--include-cues');
    }

    const result = await this.execCR8Command(command);

    return {
      content: [
        {
          type: 'text',
          text: `Rekordbox Export Result:\n${result.stdout}\n${result.stderr ? `Errors: ${result.stderr}` : ''}`,
        },
      ],
    };
  }

  async run() {
    const transport = new StdioServerTransport();
    await this.server.connect(transport);
    console.error('CR8 CLI MCP server running on stdio');
  }
}

const server = new CR8MCPServer();
server.run().catch(console.error);
