document.addEventListener('DOMContentLoaded', () => {

    const viewLauncher = document.getElementById('view-launcher');
    const viewLoading = document.getElementById('view-loading');
    const viewResults = document.getElementById('view-results');
    
    const btnScan = document.getElementById('btn-scan');
    const targetPathInput = document.getElementById('target-path');
    
    // UI Elements
    const securityGradeText = document.getElementById('security-grade-text');
    const securityScoreNum = document.getElementById('security-score-number');
    const scoreRing = document.getElementById('score-ring');
    
    const statCrit = document.getElementById('stat-crit');
    const statHigh = document.getElementById('stat-high');
    const statGit = document.getElementById('stat-git');
    
    const badgeFindings = document.getElementById('badge-findings');
    const badgeGit = document.getElementById('badge-git');
    const badgeDeploy = document.getElementById('badge-deploy');
    const badgeNetwork = document.getElementById('badge-network');
    const badgeNfr = document.getElementById('badge-nfr');

    const containers = {
        findings: document.getElementById('container-findings'),
        git: document.getElementById('container-git'),
        deploy: document.getElementById('container-deploy'),
        nfr: document.getElementById('container-nfr'),
        diagram: document.getElementById('container-diagram')
    };

    // Tab Switching
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('panel-active'));
            
            const target = e.target.getAttribute('data-tab');
            e.target.classList.add('active');
            document.getElementById('tab-' + target).classList.add('panel-active');
            
            // Re-render mermaid if diagram tab is clicked
            if (target === 'diagram') {
                mermaid.contentLoaded();
            }
        });
    });

    const startScan = async () => {
        const path = targetPathInput.value.trim();
        if (!path) return;

        // Transition to Loading
        viewLauncher.classList.add('hidden');
        viewLoading.classList.remove('hidden');

        // Loading Steps Simulation
        const steps = document.querySelectorAll('.loader-steps .step');
        let currentStep = 0;
        const interval = setInterval(() => {
            if (currentStep < steps.length - 1) {
                steps[currentStep].classList.replace('active', 'done');
                currentStep++;
                steps[currentStep].classList.replace('pending', 'active');
            }
        }, 3000);

        try {
            const resp = await fetch('/api/scan', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: path })
            });

            if (!resp.ok) {
                const data = await resp.json();
                throw new Error(data.error || 'Unknown error');
            }

            const data = await resp.json();
            clearInterval(interval);
            steps.forEach(s => { s.classList.remove('active', 'pending'); s.classList.add('done'); });
            
            setTimeout(() => renderResults(data), 500);

        } catch (err) {
            clearInterval(interval);
            alert('Scan Failed: ' + err.message);
            viewLoading.classList.add('hidden');
            viewLauncher.classList.remove('hidden');
        }
    };

    btnScan.addEventListener('click', startScan);
    targetPathInput.addEventListener('keypress', (e) => { if (e.key === 'Enter') startScan(); });

    const getSeverityColor = (sev) => {
        const u = sev.toUpperCase();
        if (u === 'CRITICAL') return 'var(--neon-red)';
        if (u === 'HIGH') return 'var(--neon-yellow)';
        if (u === 'MEDIUM') return 'var(--neon-blue)';
        return 'var(--neon-emerald)';
    };

    const createCard = (severity, title, fileLine, issueDesc, codeEvidence, advice) => {
        const col = getSeverityColor(severity);
        return `
            <div class="finding-card" style="--severity-color: ${col}">
                <div class="finding-header">
                    <h3 class="finding-title" style="color: ${col}">${title}</h3>
                    <span class="badge severity-${severity.toUpperCase()}">${severity}</span>
                </div>
                ${fileLine ? `<div class="finding-meta"><span class="finding-file"><i data-feather="file" style="width:12px"></i> ${fileLine}</span></div>` : ''}
                <div class="finding-meta" style="margin-bottom:1rem"></div>
                <p class="finding-desc">${issueDesc}</p>
                ${codeEvidence ? `<div class="code-snippet"><span class="code-highlight">${escapeHTML(codeEvidence)}</span></div>` : ''}
                ${advice ? `<div class="finding-meta" style="margin-top:1rem; color:var(--text-muted); font-size:0.85rem">💡 ${escapeHTML(advice)}</div>` : ''}
            </div>
        `;
    };

    const renderResults = (data) => {
        viewLoading.classList.add('hidden');
        viewResults.classList.remove('hidden');

        // Hero Metrics
        securityGradeText.textContent = data.security_grade;
        securityScoreNum.textContent = data.security_score + "/100";
        scoreRing.style.setProperty('--progress', data.security_score);
        
        let ringCol = 'var(--neon-emerald)';
        let glowCol = 'var(--glow-emerald)';
        if (data.security_grade === 'F' || data.security_grade === 'D') { ringCol = 'var(--neon-red)'; glowCol = 'var(--glow-red)'; }
        else if (data.security_grade === 'C') { ringCol = 'var(--neon-yellow)'; glowCol = 'var(--glow-yellow)'; }
        
        scoreRing.style.setProperty('--score-color', ringCol);
        securityGradeText.style.setProperty('--score-glow', glowCol);

        // Core Findings
        let crit = 0; let high = 0;
        let htmlCore = '';
        
        const allFindings = [...(data.findings || []), ...(data.external || [])];
        allFindings.forEach(f => {
            const s = (f.severity || '').toUpperCase();
            if (s === 'CRITICAL') crit++;
            if (s === 'HIGH') high++;
            
            const fileLine = f.file + (f.line ? `:${f.line}` : '');
            htmlCore += createCard(f.severity, f.issue || f.rule_id, fileLine, f.message, f.code_snippet, '');
        });
        
        containers.findings.innerHTML = htmlCore;
        badgeFindings.textContent = allFindings.length + ' entries';

        statCrit.textContent = crit;
        statHigh.textContent = high;

        // Git Secrets
        let htmlGit = '';
        const gitArr = data.git || [];
        gitArr.forEach(g => {
            htmlGit += createCard(g.severity, 'Secret Exposed: ' + g.issue, `${g.commit} -> ${g.file}`, g.message, g.evidence, '');
        });
        containers.git.innerHTML = htmlGit;
        badgeGit.textContent = gitArr.length + ' secrets';
        statGit.textContent = gitArr.length;

        // Deploy & Network
        let htmlDeploy = '';
        const deployArr = data.deploy || [];
        deployArr.forEach(d => {
            const fileLine = d.file + (d.line ? `:${d.line}` : '');
            htmlDeploy += createCard(d.severity, 'IaC Risk: ' + d.issue, fileLine, d.evidence, '', d.advice);
        });
        
        const netArr = data.network || [];
        netArr.forEach(n => {
            htmlDeploy += createCard(n.severity, 'Open Port: ' + n.port + ' (' + n.service + ')', n.target, n.message, '', '');
        });
        
        containers.deploy.innerHTML = htmlDeploy;
        badgeDeploy.textContent = deployArr.length + ' config issues';
        badgeNetwork.textContent = netArr.length + ' open ports';

        // NFR
        let htmlNfr = '';
        const nfrArr = data.nfr || [];
        nfrArr.forEach(n => {
            htmlNfr += createCard(n.severity, 'NFR Violation: ' + n.name, n.rule_id, n.message, '', n.advice);
        });
        containers.nfr.innerHTML = htmlNfr;
        badgeNfr.textContent = nfrArr.length + ' failures';

        // Diagram
        if (data.diagram) {
            containers.diagram.innerHTML = `<pre class="mermaid">${data.diagram}</pre>`;
        } else {
            containers.diagram.innerHTML = `<p class="text-muted">No architecture context available.</p>`;
        }

        feather.replace();
    };

    function escapeHTML(str) {
        if (!str) return '';
        var p = document.createElement("p");
        p.appendChild(document.createTextNode(str));
        return p.innerHTML;
    }
});
