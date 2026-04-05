document.addEventListener('DOMContentLoaded', () => {
    // Elements
    const queryInput = document.getElementById('query-input');
    const runBtn = document.getElementById('run-btn');
    const btnText = runBtn.querySelector('span');
    const btnLoader = runBtn.querySelector('.btn-loader');
    const errorBanner = document.getElementById('error-banner');
    const errorText = document.getElementById('error-text');
    const cacheBanner = document.getElementById('cache-banner');
    
    // Feature Sections
    const sections = {
        parser: document.getElementById('section-parser'),
        security: document.getElementById('section-security'),
        planner: document.getElementById('section-planner'),
        stats: document.getElementById('section-stats'),
        executor: document.getElementById('section-executor')
    };
    
    const connectors = document.querySelectorAll('.pipeline-connector');

    // Fetch initial stats
    fetchStats();

    runBtn.addEventListener('click', async () => {
        const query = queryInput.value.trim();
        if (!query) return;

        // Reset UI
        errorBanner.classList.add('hidden');
        cacheBanner.classList.add('hidden');
        btnText.classList.add('hidden');
        btnLoader.classList.remove('hidden');
        runBtn.disabled = true;

        // Dim all sections
        Object.values(sections).forEach(s => {
            s.classList.remove('focus', 'done');
        });
        connectors.forEach(c => c.classList.remove('active'));

        try {
            await executePipelineCascade(query);
        } catch (err) {
            showError(err.message);
        } finally {
            btnText.classList.remove('hidden');
            btnLoader.classList.add('hidden');
            runBtn.disabled = false;
        }
    });

    const delay = ms => new Promise(res => setTimeout(res, ms));

    async function executePipelineCascade(query) {
        // Fetch data
        const response = await fetch('/api/query', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query })
        });
        const data = await response.json();

        // If Cache Hit: Bypass the flow
        if (data.cache_hit) {
            cacheBanner.classList.remove('hidden');
            
            // Light everything up immediately
            connectors.forEach(c => c.classList.add('active'));
            Object.values(sections).forEach(s => s.classList.add('done'));
            
            updateResultsTab(data.results);
            updateMetrics(data.metrics);
            sections.executor.scrollIntoView({ behavior: 'smooth', block: 'center' });
            return;
        }

        // --- STEP 1: PARSER ---
        connectors[0].classList.add('active');
        sections.parser.classList.add('focus');
        sections.parser.scrollIntoView({ behavior: 'smooth', block: 'center' });
        
        await delay(600);
        
        if (!data.ast) {
            document.getElementById('parser-status').textContent = "ERROR: Could not parse query!";
            showError(data.error);
            return;
        }
        document.getElementById('parser-status').textContent = "Abstract Syntax Tree successfully generated.";
        document.getElementById('parser-status').style.color = 'var(--success)';
        document.getElementById('parser-status').style.borderColor = 'var(--success)';
        document.getElementById('ast-json').textContent = JSON.stringify(data.ast, null, 2);
        try { hljs.highlightElement(document.getElementById('ast-json')); } catch(e) {}
        
        sections.parser.classList.remove('focus');
        sections.parser.classList.add('done');

        // --- STEP 2: SECURITY ENGINE ---
        connectors[1].classList.add('active');
        sections.security.classList.add('focus');
        sections.security.scrollIntoView({ behavior: 'smooth', block: 'center' });
        
        await delay(600);
        sections.security.classList.remove('focus');
        sections.security.classList.add('done');

        if (data.decision && data.decision.Block) {
            updateSecurityTab(data.decision);
            showError(data.error);
            return;
        }
        updateSecurityTab(data.decision);

        // --- STEP 3: PLANNER & ROUTING ---
        connectors[2].classList.add('active');
        sections.planner.classList.add('focus');
        sections.planner.scrollIntoView({ behavior: 'smooth', block: 'center' });

        await delay(600);
        if (!data.plan) {
            document.getElementById('planner-status').textContent = "ERROR: Failed to route query";
            showError(data.error);
            return;
        }
        document.getElementById('planner-status').textContent = "Query split into physical database steps. NoSQL translation applied if required.";
        document.getElementById('planner-status').style.color = 'var(--success)';
        document.getElementById('planner-status').style.borderColor = 'var(--success)';
        updatePlanTab(data.plan);

        sections.planner.classList.remove('focus');
        sections.planner.classList.add('done');

        // --- STEP 4: POOL ACTIVATION ---
        // Refresh live stats
        connectors[3].classList.add('active');
        sections.stats.classList.add('focus');
        sections.stats.scrollIntoView({ behavior: 'smooth', block: 'center' });
        await fetchStats();
        
        await delay(800);
        sections.stats.classList.remove('focus');
        sections.stats.classList.add('done');

        // --- STEP 5: FINAL EXECUTION ---
        connectors[4].classList.add('active');
        sections.executor.classList.add('focus');
        sections.executor.scrollIntoView({ behavior: 'smooth', block: 'center' });

        await delay(400);
        updateResultsTab(data.results);
        updateMetrics(data.metrics);
        
        sections.executor.classList.remove('focus');
        sections.executor.classList.add('done');
    }

    function showError(msg) {
        errorText.textContent = msg;
        errorBanner.classList.remove('hidden');
    }

    function updateSecurityTab(dec) {
        const statusEl = document.getElementById('sec-status');
        const reasonEl = document.getElementById('sec-reason');
        
        statusEl.className = 'sec-status';
        if (!dec || (!dec.Block && !dec.Flag)) {
            statusEl.textContent = "Safe Query";
            statusEl.classList.add('safe');
            reasonEl.textContent = "No anomalies or injections detected.";
        } else if (dec.Block) {
            statusEl.textContent = "Query Blocked";
            statusEl.classList.add('blocked');
            reasonEl.textContent = dec.Reason;
        } else if (dec.Flag) {
            statusEl.textContent = "Query Flagged";
            statusEl.classList.add('flagged');
            reasonEl.textContent = dec.Reason;
        }
    }

    function updatePlanTab(plan) {
        const container = document.getElementById('plan-container');
        container.innerHTML = '';
        
        if (!plan.Steps || plan.Steps.length === 0) {
            container.innerHTML = '<div class="empty-state">No steps in plan.</div>';
            return;
        }

        plan.Steps.forEach(step => {
            const card = document.createElement('div');
            card.className = 'plan-card';
            
            let dependsHtml = step.DependsOn ? `<span>Depends on: ${step.DependsOn.join(', ')}</span>` : '';
            
            card.innerHTML = `
                <div class="step-id">${step.ID}</div>
                <div class="step-details">
                    <div class="step-type">${step.Type}</div>
                    <div class="step-meta">
                        ${step.Database ? `<span class="step-db">Target: ${step.Database.toUpperCase()}</span>` : ''}
                        ${dependsHtml}
                        ${step.Query ? `<div style="margin-top: 5px;"><code>${step.Query}</code></div>` : ''}
                    </div>
                </div>
            `;
            container.appendChild(card);
        });
    }

    function updateResultsTab(rows) {
        const empty = document.getElementById('results-empty');
        const table = document.getElementById('results-table');
        const thead = document.getElementById('results-thead');
        const tbody = document.getElementById('results-tbody');
        
        if (!rows || rows.length === 0) {
            empty.classList.remove('hidden');
            table.classList.add('hidden');
            empty.textContent = "Query returned 0 rows.";
            return;
        }

        empty.classList.add('hidden');
        table.classList.remove('hidden');
        
        // Headers
        thead.innerHTML = '';
        const cols = Object.keys(rows[0]);
        cols.forEach(c => {
            const th = document.createElement('th');
            th.textContent = c;
            thead.appendChild(th);
        });

        // Body
        tbody.innerHTML = '';
        rows.forEach(row => {
            const tr = document.createElement('tr');
            cols.forEach(c => {
                const td = document.createElement('td');
                td.textContent = row[c] !== null ? row[c] : 'NULL';
                tr.appendChild(td);
            });
            tbody.appendChild(tr);
        });
    }

    function updateMetrics(metrics) {
        if (!metrics) return;
        document.getElementById('metrics-bar').classList.remove('hidden');
        
        document.getElementById('m-total').textContent = (metrics.total_time_ms || 0).toFixed(2) + 'ms';
        document.getElementById('m-security').textContent = (metrics.security_time_ms || 0).toFixed(2) + 'ms';
        document.getElementById('m-parse').textContent = (metrics.parse_time_ms || 0).toFixed(2) + 'ms';
        document.getElementById('m-plan').textContent = (metrics.plan_time_ms || 0).toFixed(2) + 'ms';
        document.getElementById('m-exec').textContent = (metrics.exec_time_ms || 0).toFixed(2) + 'ms';
    }

    async function fetchStats() {
        try {
            const response = await fetch('/api/stats');
            const data = await response.json();
            
            const grid = document.getElementById('stats-grid');
            grid.innerHTML = '';

            for (const [db, stats] of Object.entries(data)) {
                const percent = (stats.active / stats.max) * 100;
                
                let dbClass = 'db-postgres';
                if (db === 'mysql') dbClass = 'db-mysql';
                if (db === 'mongodb') dbClass = 'db-mongodb';

                let fillColor = '#3b82f6';
                if (percent > 80) fillColor = '#ef4444';
                else if (percent > 60) fillColor = '#f59e0b';

                grid.innerHTML += `
                    <div class="stat-card">
                        <div class="stat-header">
                            <span class="${dbClass}">${db.toUpperCase()} Connection Pool</span>
                            <span>${stats.active} / ${stats.max} Uses</span>
                        </div>
                        <div class="gauge-bar">
                            <div class="gauge-fill" style="width: ${percent}%; background: ${fillColor}"></div>
                        </div>
                        <div class="stat-details">
                            <span>Active: ${stats.active}</span>
                            <span>Idle: ${stats.idle}</span>
                        </div>
                    </div>
                `;
            }
        } catch (e) {
            console.error(e);
        }
    }
});
