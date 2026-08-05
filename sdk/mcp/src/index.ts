#!/usr/bin/env node
/**
 * Rexec MCP server — expose sandboxes to AI agents via Model Context Protocol.
 *
 * Env:
 *   REXEC_URL   — base URL (default https://rexec.sh)
 *   REXEC_TOKEN — API bearer token (required)
 *
 * Claude Desktop / Cursor example:
 *   {
 *     "mcpServers": {
 *       "rexec": {
 *         "command": "npx",
 *         "args": ["-y", "@pipeops/rexec-mcp"],
 *         "env": { "REXEC_URL": "https://rexec.sh", "REXEC_TOKEN": "..." }
 *       }
 *     }
 *   }
 */

import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';
import { RexecClient, type ExecResult, type Sandbox } from 'pipeops-rexec';

function requireEnv(name: string): string {
  const v = process.env[name]?.trim();
  if (!v) {
    console.error(`rexec-mcp: missing required env ${name}`);
    process.exit(1);
  }
  return v;
}

function clientFromEnv(): RexecClient {
  const baseURL = (process.env.REXEC_URL || process.env.REXEC_HOST || 'https://rexec.sh').replace(
    /\/$/,
    ''
  );
  const token = requireEnv('REXEC_TOKEN');
  return new RexecClient({ baseURL, token });
}

function text(data: unknown) {
  return {
    content: [
      {
        type: 'text' as const,
        text: typeof data === 'string' ? data : JSON.stringify(data, null, 2),
      },
    ],
  };
}

function errText(err: unknown) {
  const msg = err instanceof Error ? err.message : String(err);
  return {
    content: [{ type: 'text' as const, text: `Error: ${msg}` }],
    isError: true as const,
  };
}

async function waitRunning(
  client: RexecClient,
  id: string,
  timeoutSec = 120
): Promise<Sandbox> {
  const deadline = Date.now() + timeoutSec * 1000;
  while (Date.now() < deadline) {
    const s = await client.sandboxes.get(id);
    if (s.status === 'running') return s;
    if (s.status === 'error') {
      throw new Error(`sandbox ${id} entered error state`);
    }
    await new Promise((r) => setTimeout(r, 1500));
  }
  throw new Error(`timeout waiting for sandbox ${id} to become running`);
}

async function main() {
  const client = clientFromEnv();

  const server = new McpServer({
    name: 'rexec',
    version: '1.1.0',
  });

  server.registerTool(
    'list_sandboxes',
    {
      description: 'List all Rexec sandboxes for the authenticated user.',
      inputSchema: {},
    },
    async () => {
      try {
        const list = await client.sandboxes.list();
        return text({ count: list.length, sandboxes: list });
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'create_sandbox',
    {
      description:
        'Create a Rexec sandbox. Prefer image aliases (ubuntu, debian, alpine). Optionally create from template_id or set network_mode to none for offline compute.',
      inputSchema: {
        image: z.string().optional().describe('Image alias e.g. ubuntu (omit if template_id/snapshot_id set)'),
        name: z.string().optional().describe('Optional display name'),
        template_id: z.string().optional().describe('Create from a saved template'),
        snapshot_id: z.string().optional().describe('Create from a filesystem snapshot'),
        network_mode: z
          .enum(['default', 'none', 'restricted'])
          .optional()
          .describe('default = outbound; none = no network'),
        wait_running: z
          .boolean()
          .optional()
          .describe('Poll until status is running (default true)'),
      },
    },
    async (args) => {
      try {
        const sandbox = await client.sandboxes.create({
          image: args.image,
          name: args.name,
          template_id: args.template_id,
          snapshot_id: args.snapshot_id,
          network_mode: args.network_mode,
        });
        if (args.wait_running !== false) {
          const ready = await waitRunning(client, sandbox.id);
          return text(ready);
        }
        return text(sandbox);
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'get_sandbox',
    {
      description: 'Get a sandbox by id (status, image, etc.).',
      inputSchema: {
        sandbox_id: z.string().describe('Sandbox id'),
      },
    },
    async ({ sandbox_id }) => {
      try {
        return text(await client.sandboxes.get(sandbox_id));
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'delete_sandbox',
    {
      description: 'Delete a sandbox permanently.',
      inputSchema: {
        sandbox_id: z.string(),
      },
    },
    async ({ sandbox_id }) => {
      try {
        await client.sandboxes.delete(sandbox_id);
        return text({ deleted: true, sandbox_id });
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'exec',
    {
      description:
        'Run a non-interactive shell command in a running sandbox. Returns stdout, stderr, exit_code. Prefer this over a full terminal for agents.',
      inputSchema: {
        sandbox_id: z.string(),
        command: z.string().describe('Shell command (sh -c)'),
        workdir: z.string().optional(),
        timeout_seconds: z.number().int().min(1).max(300).optional(),
      },
    },
    async ({ sandbox_id, command, workdir, timeout_seconds }) => {
      try {
        const result: ExecResult = await client.sandboxes.exec(sandbox_id, {
          command,
          workdir,
          timeout_seconds,
        });
        return text(result);
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'list_files',
    {
      description: 'List files in a directory inside a sandbox.',
      inputSchema: {
        sandbox_id: z.string(),
        path: z.string().optional().describe('Directory path (default /)'),
      },
    },
    async ({ sandbox_id, path }) => {
      try {
        const files = await client.files.list(sandbox_id, path ?? '/');
        return text(files);
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'mkdir',
    {
      description: 'Create a directory inside a sandbox.',
      inputSchema: {
        sandbox_id: z.string(),
        path: z.string(),
      },
    },
    async ({ sandbox_id, path }) => {
      try {
        await client.files.mkdir(sandbox_id, path);
        return text({ ok: true, path });
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'list_templates',
    {
      description: 'List saved sandbox templates (committed images).',
      inputSchema: {},
    },
    async () => {
      try {
        const templates = await client.sandboxes.listTemplates();
        return text({ count: templates.length, templates });
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'create_template',
    {
      description:
        'Commit a running sandbox into a reusable template image. Then create new sandboxes with template_id.',
      inputSchema: {
        name: z.string().describe('Template name'),
        from_sandbox_id: z.string().describe('Running sandbox id to commit'),
        description: z.string().optional(),
      },
    },
    async ({ name, from_sandbox_id, description }) => {
      try {
        const tpl = await client.sandboxes.createTemplate({
          name,
          from_sandbox_id,
          description,
        });
        return text(tpl);
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'delete_template',
    {
      description: 'Delete a sandbox template by id.',
      inputSchema: {
        template_id: z.string(),
      },
    },
    async ({ template_id }) => {
      try {
        await client.sandboxes.deleteTemplate(template_id);
        return text({ deleted: true, template_id });
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'wait_running',
    {
      description: 'Poll until a sandbox status is running (or error/timeout).',
      inputSchema: {
        sandbox_id: z.string(),
        timeout_seconds: z.number().int().min(5).max(600).optional(),
      },
    },
    async ({ sandbox_id, timeout_seconds }) => {
      try {
        const s = await waitRunning(client, sandbox_id, timeout_seconds ?? 120);
        return text(s);
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'snapshot_sandbox',
    {
      description:
        'Create a point-in-time filesystem snapshot of a sandbox (docker commit). Use snapshot_id later to spawn clones.',
      inputSchema: {
        sandbox_id: z.string(),
        name: z.string().optional(),
        description: z.string().optional(),
      },
    },
    async ({ sandbox_id, name, description }) => {
      try {
        const snap = await client.sandboxes.snapshot(sandbox_id, { name, description });
        return text(snap);
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'list_snapshots',
    {
      description: 'List filesystem snapshots for the current user.',
      inputSchema: {},
    },
    async () => {
      try {
        const snapshots = await client.sandboxes.listSnapshots();
        return text({ count: snapshots.length, snapshots });
      } catch (e) {
        return errText(e);
      }
    }
  );

  server.registerTool(
    'fork_sandbox',
    {
      description:
        'Fork a sandbox: commit its filesystem and create a new running sandbox from that image.',
      inputSchema: {
        sandbox_id: z.string(),
        name: z.string().optional(),
        network_mode: z.enum(['default', 'none', 'restricted']).optional(),
        save_snapshot: z.boolean().optional(),
      },
    },
    async ({ sandbox_id, name, network_mode, save_snapshot }) => {
      try {
        const fork = await client.sandboxes.fork(sandbox_id, {
          name,
          network_mode,
          save_snapshot,
        });
        return text(fork);
      } catch (e) {
        return errText(e);
      }
    }
  );

  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((err) => {
  console.error('rexec-mcp failed:', err);
  process.exit(1);
});
