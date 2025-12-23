#!/usr/bin/env node

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import { existsSync, readdirSync, readFileSync, statSync, writeFileSync } from 'fs';
import { glob } from 'glob';
import matter from 'gray-matter';
import { marked } from 'marked';
import { dirname, join, relative } from 'path';
import yaml from 'yaml';

const server = new Server(
  {
    name: 'aftrs-wiki-server',
    version: '0.1.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Base wiki path - adjust as needed
const WIKI_BASE_PATH = '/home/hg/Docs/aftrs-void/aftrs_wiki';

server.setRequestHandler(ListToolsRequestSchema, async () => {
  return {
    tools: [
      {
        name: 'list_wiki_categories',
        description: 'List all wiki categories and their document counts',
        inputSchema: {
          type: 'object',
          properties: {},
        },
      },
      {
        name: 'list_documents',
        description: 'List documents in a specific category or all documents',
        inputSchema: {
          type: 'object',
          properties: {
            category: {
              type: 'string',
              description: 'Wiki category (infrastructure, projects, patterns, roadmap, miscellaneous)',
            },
            includeMetadata: {
              type: 'boolean',
              description: 'Include document metadata in results',
              default: false,
            },
          },
        },
      },
      {
        name: 'read_document',
        description: 'Read a specific wiki document with optional format conversion',
        inputSchema: {
          type: 'object',
          properties: {
            path: {
              type: 'string',
              description: 'Relative path to document within wiki (e.g., "projects/aftrs_cli.md")',
            },
            format: {
              type: 'string',
              enum: ['raw', 'parsed', 'html'],
              description: 'Output format - raw markdown, parsed with metadata, or HTML',
              default: 'parsed',
            },
          },
          required: ['path'],
        },
      },
      {
        name: 'search_wiki',
        description: 'Search across all wiki documents for content',
        inputSchema: {
          type: 'object',
          properties: {
            query: {
              type: 'string',
              description: 'Search query string',
            },
            category: {
              type: 'string',
              description: 'Limit search to specific category (optional)',
            },
            caseSensitive: {
              type: 'boolean',
              description: 'Case sensitive search',
              default: false,
            },
          },
          required: ['query'],
        },
      },
      {
        name: 'create_document',
        description: 'Create a new wiki document',
        inputSchema: {
          type: 'object',
          properties: {
            path: {
              type: 'string',
              description: 'Relative path for new document (e.g., "projects/new_project.md")',
            },
            content: {
              type: 'string',
              description: 'Document content (markdown with optional frontmatter)',
            },
            metadata: {
              type: 'object',
              description: 'Optional metadata object to include in frontmatter',
            },
          },
          required: ['path', 'content'],
        },
      },
      {
        name: 'update_document',
        description: 'Update an existing wiki document',
        inputSchema: {
          type: 'object',
          properties: {
            path: {
              type: 'string',
              description: 'Relative path to document to update',
            },
            content: {
              type: 'string',
              description: 'New document content (markdown with optional frontmatter)',
            },
            mergeMetadata: {
              type: 'boolean',
              description: 'Merge new metadata with existing instead of replacing',
              default: true,
            },
          },
          required: ['path', 'content'],
        },
      },
      {
        name: 'get_project_inventory',
        description: 'Read and parse project inventory TSV files',
        inputSchema: {
          type: 'object',
          properties: {
            project: {
              type: 'string',
              description: 'Project name (aftrs-ai, console-hax, org-admin, secretstudios, hairglasses)',
            },
          },
          required: ['project'],
        },
      },
      {
        name: 'cross_reference_search',
        description: 'Find cross-references and links between wiki documents',
        inputSchema: {
          type: 'object',
          properties: {
            document: {
              type: 'string',
              description: 'Document path to find references for',
            },
            searchType: {
              type: 'string',
              enum: ['incoming', 'outgoing', 'both'],
              description: 'Type of references to find',
              default: 'both',
            },
          },
          required: ['document'],
        },
      },
      {
        name: 'get_wiki_statistics',
        description: 'Get comprehensive statistics about the wiki content',
        inputSchema: {
          type: 'object',
          properties: {
            detailed: {
              type: 'boolean',
              description: 'Include detailed per-category statistics',
              default: false,
            },
          },
        },
      },
      {
        name: 'extract_todos_and_issues',
        description: 'Extract TODO items and issues from wiki documents',
        inputSchema: {
          type: 'object',
          properties: {
            category: {
              type: 'string',
              description: 'Limit to specific category (optional)',
            },
            includeCompleted: {
              type: 'boolean',
              description: 'Include completed TODO items',
              default: false,
            },
          },
        },
      },
    ],
  };
});

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  try {
    switch (name) {
      case 'list_wiki_categories': {
        const categories = readdirSync(WIKI_BASE_PATH, { withFileTypes: true })
          .filter(dirent => dirent.isDirectory())
          .map(dirent => {
            const categoryPath = join(WIKI_BASE_PATH, dirent.name);
            const files = glob.sync('**/*.md', { cwd: categoryPath });
            return {
              name: dirent.name,
              documentCount: files.length,
              path: dirent.name,
            };
          });

        return {
          content: [
            {
              type: 'text',
              text: `Wiki Categories:\n\n${JSON.stringify(categories, null, 2)}`,
            },
          ],
        };
      }

      case 'list_documents': {
        let searchPattern: string;
        let searchPath: string;

        if (args?.category) {
          searchPath = join(WIKI_BASE_PATH, args.category);
          searchPattern = '**/*.md';
        } else {
          searchPath = WIKI_BASE_PATH;
          searchPattern = '**/*.md';
        }

        const files = glob.sync(searchPattern, { cwd: searchPath });
        const documents = files.map(file => {
          const fullPath = join(searchPath, file);
          const relativePath = relative(WIKI_BASE_PATH, fullPath);
          const stats = statSync(fullPath);

          let metadata = null;
          if (args?.includeMetadata) {
            try {
              const content = readFileSync(fullPath, 'utf8');
              const parsed = matter(content);
              metadata = parsed.data;
            } catch (error) {
              // Ignore metadata parsing errors
            }
          }

          return {
            path: relativePath,
            name: file,
            size: stats.size,
            modified: stats.mtime.toISOString(),
            metadata,
          };
        });

        return {
          content: [
            {
              type: 'text',
              text: `Found ${documents.length} documents:\n\n${JSON.stringify(documents, null, 2)}`,
            },
          ],
        };
      }

      case 'read_document': {
        const fullPath = join(WIKI_BASE_PATH, args.path);

        if (!existsSync(fullPath)) {
          throw new Error(`Document not found: ${args.path}`);
        }

        const content = readFileSync(fullPath, 'utf8');

        switch (args.format) {
          case 'raw':
            return {
              content: [
                {
                  type: 'text',
                  text: content,
                },
              ],
            };

          case 'html':
            const parsed = matter(content);
            const html = await marked(parsed.content);
            return {
              content: [
                {
                  type: 'text',
                  text: `Metadata: ${JSON.stringify(parsed.data, null, 2)}\n\nContent:\n${html}`,
                },
              ],
            };

          case 'parsed':
          default:
            const parsedDoc = matter(content);
            return {
              content: [
                {
                  type: 'text',
                  text: `Document: ${args.path}\n\nMetadata:\n${JSON.stringify(parsedDoc.data, null, 2)}\n\nContent:\n${parsedDoc.content}`,
                },
              ],
            };
        }
      }

      case 'search_wiki': {
        const searchPattern = '**/*.md';
        let searchPath = WIKI_BASE_PATH;

        if (args.category) {
          searchPath = join(WIKI_BASE_PATH, args.category);
        }

        const files = glob.sync(searchPattern, { cwd: searchPath });
        const results: any[] = [];
        const query = args.caseSensitive ? args.query : args.query.toLowerCase();

        for (const file of files) {
          const fullPath = join(searchPath, file);
          const content = readFileSync(fullPath, 'utf8');
          const searchContent = args.caseSensitive ? content : content.toLowerCase();

          if (searchContent.includes(query)) {
            const lines = content.split('\n');
            const matchingLines: { lineNumber: number; text: string }[] = [];

            lines.forEach((line, index) => {
              const searchLine = args.caseSensitive ? line : line.toLowerCase();
              if (searchLine.includes(query)) {
                matchingLines.push({
                  lineNumber: index + 1,
                  text: line.trim(),
                });
              }
            });

            results.push({
              file: relative(WIKI_BASE_PATH, fullPath),
              matches: matchingLines.length,
              matchingLines,
            });
          }
        }

        return {
          content: [
            {
              type: 'text',
              text: `Search results for "${args.query}":\n\n${JSON.stringify(results, null, 2)}`,
            },
          ],
        };
      }

      case 'create_document': {
        const fullPath = join(WIKI_BASE_PATH, args.path);

        if (existsSync(fullPath)) {
          throw new Error(`Document already exists: ${args.path}`);
        }

        let content = args.content;

        // Add frontmatter if metadata provided
        if (args.metadata) {
          const frontmatter = yaml.stringify(args.metadata);
          content = `---\n${frontmatter}---\n\n${args.content}`;
        }

        // Ensure directory exists
        const dir = dirname(fullPath);
        if (!existsSync(dir)) {
          throw new Error(`Directory does not exist: ${relative(WIKI_BASE_PATH, dir)}`);
        }

        writeFileSync(fullPath, content, 'utf8');

        return {
          content: [
            {
              type: 'text',
              text: `Successfully created document: ${args.path}`,
            },
          ],
        };
      }

      case 'update_document': {
        const fullPath = join(WIKI_BASE_PATH, args.path);

        if (!existsSync(fullPath)) {
          throw new Error(`Document not found: ${args.path}`);
        }

        let content = args.content;

        if (args.mergeMetadata) {
          const existingContent = readFileSync(fullPath, 'utf8');
          const existingParsed = matter(existingContent);
          const newParsed = matter(args.content);

          // Merge metadata
          const mergedMetadata = { ...existingParsed.data, ...newParsed.data };

          if (Object.keys(mergedMetadata).length > 0) {
            const frontmatter = yaml.stringify(mergedMetadata);
            content = `---\n${frontmatter}---\n\n${newParsed.content}`;
          }
        }

        writeFileSync(fullPath, content, 'utf8');

        return {
          content: [
            {
              type: 'text',
              text: `Successfully updated document: ${args.path}`,
            },
          ],
        };
      }

      case 'get_project_inventory': {
        const inventoryFile = `${args.project}-inventory.tsv`;
        const inventoryPath = join(WIKI_BASE_PATH, 'projects', inventoryFile);

        if (!existsSync(inventoryPath)) {
          throw new Error(`Inventory file not found: ${inventoryFile}`);
        }

        const content = readFileSync(inventoryPath, 'utf8');
        const lines = content.split('\n').filter(line => line.trim());
        const headers = lines[0].split('\t');
        const data = lines.slice(1).map(line => {
          const values = line.split('\t');
          const row: any = {};
          headers.forEach((header, index) => {
            row[header] = values[index] || '';
          });
          return row;
        });

        return {
          content: [
            {
              type: 'text',
              text: `Project Inventory: ${args.project}\n\n${JSON.stringify(data, null, 2)}`,
            },
          ],
        };
      }

      case 'cross_reference_search': {
        const documentName = args.document.replace(/\.md$/, '');
        const files = glob.sync('**/*.md', { cwd: WIKI_BASE_PATH });

        const incomingRefs: string[] = [];
        const outgoingRefs: string[] = [];

        // Find incoming references (other docs that mention this doc)
        if (args.searchType === 'incoming' || args.searchType === 'both') {
          for (const file of files) {
            if (file === args.document) continue;

            const fullPath = join(WIKI_BASE_PATH, file);
            const content = readFileSync(fullPath, 'utf8');

            if (content.includes(documentName) || content.includes(args.document)) {
              incomingRefs.push(file);
            }
          }
        }

        // Find outgoing references (docs this doc mentions)
        if (args.searchType === 'outgoing' || args.searchType === 'both') {
          const docPath = join(WIKI_BASE_PATH, args.document);
          if (existsSync(docPath)) {
            const content = readFileSync(docPath, 'utf8');

            for (const file of files) {
              if (file === args.document) continue;

              const fileName = file.replace(/\.md$/, '');
              if (content.includes(fileName) || content.includes(file)) {
                outgoingRefs.push(file);
              }
            }
          }
        }

        return {
          content: [
            {
              type: 'text',
              text: `Cross-references for "${args.document}":\n\nIncoming: ${JSON.stringify(incomingRefs, null, 2)}\n\nOutgoing: ${JSON.stringify(outgoingRefs, null, 2)}`,
            },
          ],
        };
      }

      case 'get_wiki_statistics': {
        const files = glob.sync('**/*.md', { cwd: WIKI_BASE_PATH });
        let totalSize = 0;
        let totalWords = 0;
        const categoryStats: any = {};

        for (const file of files) {
          const fullPath = join(WIKI_BASE_PATH, file);
          const content = readFileSync(fullPath, 'utf8');
          const stats = statSync(fullPath);
          const wordCount = content.split(/\s+/).length;

          totalSize += stats.size;
          totalWords += wordCount;

          const category = file.split('/')[0];
          if (!categoryStats[category]) {
            categoryStats[category] = {
              files: 0,
              size: 0,
              words: 0,
            };
          }
          categoryStats[category].files++;
          categoryStats[category].size += stats.size;
          categoryStats[category].words += wordCount;
        }

        const result = {
          totalFiles: files.length,
          totalSize,
          totalWords,
          categories: Object.keys(categoryStats).length,
        };

        if (args.detailed) {
          (result as any).categoryDetails = categoryStats;
        }

        return {
          content: [
            {
              type: 'text',
              text: `Wiki Statistics:\n\n${JSON.stringify(result, null, 2)}`,
            },
          ],
        };
      }

      case 'extract_todos_and_issues': {
        let searchPath = WIKI_BASE_PATH;
        if (args.category) {
          searchPath = join(WIKI_BASE_PATH, args.category);
        }

        const files = glob.sync('**/*.md', { cwd: searchPath });
        const todos: any[] = [];
        const issues: any[] = [];

        const todoRegex = /^[\s]*[-*]\s*\[([x\s])\]\s*(.+)$/gim;
        const issueRegex = /(?:issue|problem|bug|error|fix|todo):\s*(.+)/gi;

        for (const file of files) {
          const fullPath = join(searchPath, file);
          const content = readFileSync(fullPath, 'utf8');
          const lines = content.split('\n');

          // Extract TODOs
          lines.forEach((line, index) => {
            const match = todoRegex.exec(line);
            if (match) {
              const isCompleted = match[1].toLowerCase() === 'x';
              if (args.includeCompleted || !isCompleted) {
                todos.push({
                  file: relative(WIKI_BASE_PATH, fullPath),
                  lineNumber: index + 1,
                  completed: isCompleted,
                  text: match[2].trim(),
                });
              }
            }
            todoRegex.lastIndex = 0;

            // Extract issues
            const issueMatch = issueRegex.exec(line);
            if (issueMatch) {
              issues.push({
                file: relative(WIKI_BASE_PATH, fullPath),
                lineNumber: index + 1,
                text: issueMatch[1].trim(),
              });
            }
            issueRegex.lastIndex = 0;
          });
        }

        return {
          content: [
            {
              type: 'text',
              text: `Extracted TODOs and Issues:\n\nTODOs (${todos.length}):\n${JSON.stringify(todos, null, 2)}\n\nIssues (${issues.length}):\n${JSON.stringify(issues, null, 2)}`,
            },
          ],
        };
      }

      default:
        throw new Error(`Unknown tool: ${name}`);
    }
  } catch (error) {
    return {
      content: [
        {
          type: 'text',
          text: `Error executing ${name}: ${error.message}`,
        },
      ],
      isError: true,
    };
  }
});

async function runServer() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('AFTRS Wiki MCP Server running on stdio');
}

runServer().catch(console.error);
