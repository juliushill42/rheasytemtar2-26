// Copyright (c) 2026 Julius Cameron Hill. All rights reserved.
import { serve } from "https://deno.land/std@0.168.0/http/server.ts";

const DASHBOARD_PORT = Deno.env.get("DASHBOARD_PORT") || "9190";
const ORCHESTRA_URL = Deno.env.get("ORCHESTRA_URL") || "http://localhost:9100";
const BRAIN_URL = Deno.env.get("BRAIN_URL") || "http://localhost:9101";
const CLONING_URL = Deno.env.get("CLONING_URL") || "http://localhost:9102";
const SANCTUARY_URL = Deno.env.get("SANCTUARY_URL") || "http://localhost:9103";
const LEDGER_URL = Deno.env.get("LEDGER_URL") || "http://localhost:9104";

interface ServiceStatus {
  name: string;
  url: string;
  healthy: boolean;
  latency: number;
  data?: any;
}

async function checkService(name: string, url: string): Promise<ServiceStatus> {
  const start = Date.now();
  try {
    const res = await fetch(`${url}/health`, { signal: AbortSignal.timeout(3000) });
    const data = await res.json();
    return {
      name,
      url,
      healthy: res.ok,
      latency: Date.now() - start,
      data
    };
  } catch {
    return {
      name,
      url,
      healthy: false,
      latency: Date.now() - start
    };
  }
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>RHEA Glass Dashboard</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-950 text-gray-100 min-h-screen">
  <div class="max-w-7xl mx-auto p-6">
    <header class="mb-8">
      <h1 class="text-4xl font-bold bg-gradient-to-r from-cyan-400 to-blue-500 bg-clip-text text-transparent">
        RHEA Glass Dashboard
      </h1>
      <p class="text-gray-400 mt-2">Distributed System Orchestration Monitor</p>
    </header>

    <div id="status" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
      <div class="animate-pulse bg-gray-800 rounded-lg p-6 h-32"></div>
      <div class="animate-pulse bg-gray-800 rounded-lg p-6 h-32"></div>
      <div class="animate-pulse bg-gray-800 rounded-lg p-6 h-32"></div>
    </div>

    <div class="bg-gray-900 rounded-lg p-6 border border-gray-800">
      <h2 class="text-xl font-semibold mb-4">System Architecture</h2>
      <div class="space-y-3 text-sm">
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-cyan-500"></div>
          <span class="text-gray-300"><span class="font-mono text-cyan-400">Orchestra</span> - Traffic coordination & routing</span>
        </div>
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-purple-500"></div>
          <span class="text-gray-300"><span class="font-mono text-purple-400">Brain</span> - Policy engine & decision making</span>
        </div>
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-green-500"></div>
          <span class="text-gray-300"><span class="font-mono text-green-400">Cloning</span> - Blueprint management & replication</span>
        </div>
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-amber-500"></div>
          <span class="text-gray-300"><span class="font-mono text-amber-400">Sanctuary</span> - Quarantine & threat isolation</span>
        </div>
        <div class="flex items-center gap-3">
          <div class="w-3 h-3 rounded-full bg-blue-500"></div>
          <span class="text-gray-300"><span class="font-mono text-blue-400">Ledger</span> - Immutable audit log</span>
        </div>
      </div>
    </div>
  </div>

  <script>
    const services = [
      { name: 'Orchestra', url: '/api/status/orchestra', color: 'cyan' },
      { name: 'Brain', url: '/api/status/brain', color: 'purple' },
      { name: 'Cloning', url: '/api/status/cloning', color: 'green' },
      { name: 'Sanctuary', url: '/api/status/sanctuary', color: 'amber' },
      { name: 'Ledger', url: '/api/status/ledger', color: 'blue' }
    ];

    async function updateStatus() {
      const results = await Promise.all(
        services.map(s => fetch(s.url).then(r => r.json()))
      );

      const container = document.getElementById('status');
      container.innerHTML = results.map((status, i) => {
        const svc = services[i];
        const healthColor = status.healthy ? 'green' : 'red';
        const bgColor = status.healthy ? 'gray-800' : 'gray-800/50';
        
        return \`
          <div class="bg-\${bgColor} rounded-lg p-6 border border-gray-700 hover:border-\${svc.color}-500 transition-colors">
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-lg font-semibold text-\${svc.color}-400">\${svc.name}</h3>
              <div class="flex items-center gap-2">
                <div class="w-2 h-2 rounded-full bg-\${healthColor}-500 animate-pulse"></div>
                <span class="text-xs text-gray-400">\${status.latency}ms</span>
              </div>
            </div>
            <div class="space-y-1 text-sm text-gray-400">
              <div>Status: <span class="text-\${healthColor}-400 font-medium">\${status.healthy ? 'Healthy' : 'Degraded'}</span></div>
              \${status.data ? Object.entries(status.data)
                .filter(([k]) => k !== 'status')
                .map(([k, v]) => \`<div>\${k}: <span class="text-gray-200">\${v}</span></div>\`)
                .join('') : ''}
            </div>
          </div>
        \`;
      }).join('');
    }

    updateStatus();
    setInterval(updateStatus, 5000);
  </script>
</body>
</html>
`;

async function handler(req: Request): Promise<Response> {
  const url = new URL(req.url);

  if (url.pathname === "/") {
    return new Response(htmlTemplate, {
      headers: { "content-type": "text/html" }
    });
  }

  const statusMap: Record<string, string> = {
    "/api/status/orchestra": ORCHESTRA_URL,
    "/api/status/brain": BRAIN_URL,
    "/api/status/cloning": CLONING_URL,
    "/api/status/sanctuary": SANCTUARY_URL,
    "/api/status/ledger": LEDGER_URL
  };

  if (url.pathname in statusMap) {
    const serviceName = url.pathname.split("/").pop() || "";
    const serviceUrl = statusMap[url.pathname];
    const status = await checkService(serviceName, serviceUrl);
    
    return new Response(JSON.stringify(status), {
      headers: { "content-type": "application/json" }
    });
  }

  return new Response("Not Found", { status: 404 });
}

console.log(`[GLASS] Dashboard starting on port ${DASHBOARD_PORT}`);
await serve(handler, { port: parseInt(DASHBOARD_PORT) });
