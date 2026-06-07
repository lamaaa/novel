const API = '/api';

// --- Router ---
function navigate(path) {
    history.pushState(null, '', path);
    handleRoute();
}

window.addEventListener('popstate', handleRoute);

function handleRoute() {
    const path = window.location.pathname;
    const app = document.getElementById('app');

    if (path === '/' || path === '') {
        renderNovelList();
    } else if (path.match(/^\/novel\/\d+$/)) {
        const id = path.split('/')[2];
        renderNovelDetail(id);
    } else if (path.match(/^\/chapter\/\d+$/)) {
        const id = path.split('/')[2];
        renderChapterReading(id);
    } else {
        app.innerHTML = '<div class="empty-state"><div class="icon">🔍</div><p>页面不存在</p></div>';
    }
}

// --- API helpers ---
async function api(path, options = {}) {
    const res = await fetch(API + path, {
        headers: { 'Content-Type': 'application/json', ...options.headers },
        ...options
    });
    return res.json();
}

// --- Novel List ---
async function renderNovelList() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <div class="page-header">
            <h2>📚 我的书架</h2>
        </div>
        <div class="search-bar">
            <input type="text" id="search-keyword" placeholder="搜索小说标题或作者..." onkeyup="if(event.key==='Enter')searchNovels()">
            <button class="btn btn-primary" onclick="searchNovels()">搜索</button>
        </div>
        <div id="novel-list"></div>
    `;
    await loadNovels();
}

async function loadNovels(keyword = '') {
    const params = new URLSearchParams();
    if (keyword) params.set('keyword', keyword);
    const res = await api('/novels?' + params.toString());
    const list = res.data?.list || [];
    const container = document.getElementById('novel-list');
    if (!list.length) {
        container.innerHTML = '<div class="empty-state"><div class="icon">📖</div><p>还没有小说</p></div>';
        return;
    }
    container.innerHTML = `<div class="novel-grid">${list.map(n => `
        <div class="card novel-card card-clickable" onclick="navigate('/novel/${n.id}')">
            <div class="cover">${n.cover_url ? `<img src="${n.cover_url}">` : '📖'}</div>
            <h3>${esc(n.title)}</h3>
            <div class="meta">${esc(n.author || '佚名')} · ${n.chapter_count || 0}章</div>
            <span class="status-tag status-${n.status}">${n.status === 0 ? '连载中' : '已完结'}</span>
        </div>
    `).join('')}</div>`;
}

function searchNovels() {
    const keyword = document.getElementById('search-keyword').value;
    loadNovels(keyword);
}

// --- Novel Detail ---
async function renderNovelDetail(id) {
    const app = document.getElementById('app');
    const res = await api('/novels/' + id);
    const novel = res.data;
    if (!novel) { app.innerHTML = '<p>小说不存在</p>'; return; }

    app.innerHTML = `
        <div class="back-link" onclick="navigate('/')">← 返回书架</div>
        <div class="card" style="display:flex;gap:24px;align-items:flex-start;">
            <div class="novel-detail-cover">
                ${novel.cover_url ? `<img src="${novel.cover_url}">` : '📖'}
            </div>
            <div style="flex:1">
                <h2 style="margin-bottom:8px">${esc(novel.title)}</h2>
                <p style="color:var(--text-secondary);margin-bottom:4px">${esc(novel.author || '佚名')} · ${novel.chapter_count || 0}章 · <span class="status-tag status-${novel.status}">${novel.status===0?'连载中':'已完结'}</span></p>
                <p style="margin:12px 0;color:var(--text-secondary)">${esc(novel.description || '暂无简介')}</p>
            </div>
        </div>
        <div class="tabs" id="detail-tabs">
            <div class="tab active" onclick="switchTab('chapters')">目录</div>
            <div class="tab" onclick="switchTab('characters')">人物</div>
            <div class="tab" onclick="switchTab('worldviews')">世界观</div>
            <div class="tab" onclick="switchTab('foreshadowings')">伏笔</div>
            <div class="tab" onclick="switchTab('memory')">长期记忆</div>
        </div>
        <div id="tab-content"></div>
    `;
    currentNovelId = id;
    switchTab('chapters');
}

let currentNovelId = null;
let currentTab = 'chapters';

async function switchTab(tab) {
    currentTab = tab;
    document.querySelectorAll('#detail-tabs .tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('#detail-tabs .tab').forEach(t => { if (t.textContent.includes(tabName(tab))) t.classList.add('active'); });

    const container = document.getElementById('tab-content');
    if (tab === 'chapters') await loadChapters();
    else if (tab === 'characters') await loadCharacters();
    else if (tab === 'worldviews') await loadWorldviews();
    else if (tab === 'foreshadowings') await loadForeshadowings();
    else if (tab === 'memory') await loadMemory();
}

function tabName(t) {
    return { chapters: '目录', characters: '人物', worldviews: '世界观', foreshadowings: '伏笔', memory: '长期记忆' }[t] || t;
}

// --- Chapters ---
async function loadChapters() {
    const res = await api('/novels/' + currentNovelId + '/chapters');
    const chapters = res.data || [];
    const container = document.getElementById('tab-content');
    if (!chapters.length) {
        container.innerHTML = '<div class="empty-state"><div class="icon">📝</div><p>还没有章节</p></div>';
        return;
    }
    container.innerHTML = `
        <ul class="chapter-list card" style="padding:0">
        ${chapters.map(ch => `
            <li onclick="navigate('/chapter/${ch.id}')">
                <div><span class="title">${ch.chapter_order}. ${esc(ch.title)}</span> ${ch.status===0?'<span class="draft-tag">草稿</span>':''}</div>
                <div class="meta">${ch.word_count}字</div>
            </li>
        `).join('')}
        </ul>
    `;
}

// --- Chapter Reading ---
async function renderChapterReading(id) {
    const app = document.getElementById('app');
    const res = await api('/chapters/' + id);
    const data = res.data;
    if (!data) { app.innerHTML = '<p>章节不存在</p>'; return; }
    const ch = data.chapter;

    app.innerHTML = `
        <div class="back-link" onclick="navigate('/novel/${ch.novel_id}')">← 返回目录</div>
        <div class="reading-mode">
            <div class="chapter-title">${esc(ch.title)}</div>
            <div class="chapter-content">${esc(ch.content)}</div>
            <div class="reading-nav">
                ${data.prev_id ? `<button class="btn btn-outline" onclick="navigate('/chapter/${data.prev_id}')">上一章</button>` : '<span></span>'}
                ${data.next_id ? `<button class="btn btn-outline" onclick="navigate('/chapter/${data.next_id}')">下一章</button>` : '<span></span>'}
            </div>
        </div>
    `;
    window.scrollTo(0, 0);
}

// --- Characters ---
async function loadCharacters() {
    const res = await api('/novels/' + currentNovelId + '/characters');
    const characters = res.data || [];
    const container = document.getElementById('tab-content');
    if (!characters.length) {
        container.innerHTML = '<div class="empty-state"><div class="icon">👤</div><p>还没有人物</p></div>';
        return;
    }
    container.innerHTML = `
        <div class="character-grid">
        ${characters.map(ch => `
            <div class="card character-card card-clickable" onclick="showCharacterDetail(${ch.id})">
                <div class="avatar">${ch.avatar_url ? `<img src="${ch.avatar_url}" style="width:100%;height:100%;border-radius:50%;object-fit:cover">` : '👤'}</div>
                <div class="info">
                    <h4>${esc(ch.name)}${ch.alias ? ` (${esc(ch.alias)})` : ''}</h4>
                    <p>${genderText(ch.gender)}${ch.age ? ' · ' + ageText(ch.age) : ''}</p>
                    <p style="margin-top:4px">${esc((ch.description||'').substring(0,60))}${(ch.description||'').length>60?'...':''}</p>
                </div>
            </div>
        `).join('')}
        </div>
    `;
}

function genderText(g) { return {0:'未知',1:'男',2:'女'}[g]||'未知'; }
function ageText(age) {
    if (!age) return '';
    const s = String(age);
    if (s.includes('岁')) return s;
    return s + '岁';
}

async function showCharacterDetail(id) {
    const res = await api('/characters/' + id);
    const ch = res.data;
    if (!ch) return;
    openModal(`
        <h3 style="margin-bottom:12px">${esc(ch.name)}${ch.alias?' ('+esc(ch.alias)+')':''}</h3>
        <p><strong>性别：</strong>${genderText(ch.gender)}${ch.age?' · '+ch.age:''}</p>
        ${ch.description?`<div style="margin-top:12px"><strong>描述</strong><p>${esc(ch.description)}</p></div>`:''}
        ${ch.personality?`<div style="margin-top:12px"><strong>性格</strong><p>${esc(ch.personality)}</p></div>`:''}
        ${ch.background?`<div style="margin-top:12px"><strong>背景</strong><p>${esc(ch.background)}</p></div>`:''}
        <div style="margin-top:16px;text-align:right"><button class="btn btn-outline" onclick="closeModal()">关闭</button></div>
    `);
}

// --- Worldviews ---
async function loadWorldviews() {
    const res = await api('/novels/' + currentNovelId + '/worldviews');
    const data = res.data || {};
    const worldviews = data.list || [];
    const categories = data.categories || [];
    const container = document.getElementById('tab-content');

    if (!worldviews.length) {
        container.innerHTML = '<div class="empty-state"><div class="icon">🌍</div><p>还没有世界观设定</p></div>';
        return;
    }

    let html = '';
    categories.forEach(cat => {
        const items = worldviews.filter(w => w.category === cat);
        html += `<div class="worldview-group"><h3>📍 ${esc(cat)}</h3>`;
        items.forEach(w => {
            html += `<div class="card" style="margin-bottom:8px">
                <div style="display:flex;justify-content:space-between;align-items:center">
                    <strong>${esc(w.title)}</strong>
                </div>
                <p style="margin-top:8px;color:var(--text-secondary);white-space:pre-wrap">${esc(w.content)}</p>
            </div>`;
        });
        html += '</div>';
    });

    container.innerHTML = html;
}

// --- Foreshadowings ---
async function loadForeshadowings() {
    const [res, chaptersRes] = await Promise.all([
        api('/novels/' + currentNovelId + '/foreshadowings'),
        api('/novels/' + currentNovelId + '/chapters')
    ]);
    const list = res.data || [];
    const chapters = chaptersRes.data || [];
    const container = document.getElementById('tab-content');

    if (!list.length) {
        container.innerHTML = '<div class="empty-state"><div class="icon">🎯</div><p>还没有伏笔</p></div>';
        return;
    }

    const statusMap = {0:'已埋设',1:'已回收',2:'已放弃'};
    container.innerHTML = `
        ${list.map(f => `
            <div class="card">
                <div style="display:flex;justify-content:space-between;align-items:flex-start">
                    <div>
                        <strong>${esc(f.title)}</strong>
                        <span class="foreshadowing-status fs-status-${f.status}">${statusMap[f.status]}</span>
                        <div class="importance-bar">${[1,2,3,4,5].map(i=>`<div class="dot ${i<=f.importance?'filled':''}"></div>`).join('')}</div>
                    </div>
                </div>
                ${f.description?`<p style="margin-top:8px;color:var(--text-secondary)">${esc(f.description)}</p>`:''}
                <div style="margin-top:8px;font-size:13px;color:var(--text-secondary)">
                    ${f.planted_chapter_title?`埋设：${esc(f.planted_chapter_title)}`:''}
                    ${f.resolved_chapter_title?` → 回收：${esc(f.resolved_chapter_title)}`:''}
                </div>
            </div>
        `).join('')}
    `;
}

// --- Long-term Memory ---
async function loadMemory() {
    const container = document.getElementById('tab-content');
    container.innerHTML = '<div class="empty-state"><div class="icon">⏳</div><p>正在读取长期记忆...</p></div>';
    const res = await api('/novels/' + currentNovelId + '/memory?limit=50');
    if (res.code !== 0) {
        container.innerHTML = `
            <div class="memory-warning">
                <strong>长期记忆暂不可用</strong>
                <p>${esc(res.message || '请确认记忆表已创建')}</p>
                <p class="muted">需要先在数据库执行 <code>novel/sql/memory.sql</code>，再让 agent 补录章节记忆。</p>
            </div>
        `;
        return;
    }
    renderMemory(res.data || {});
}

function renderMemory(data) {
    const summaries = data.summaries || [];
    const characters = data.characters || [];
    const memories = data.memories || [];
    const timeline = data.timeline || [];
    const counts = data.counts || {};
    const container = document.getElementById('tab-content');
    const total = (counts.summaries || 0) + (counts.characters || 0) + (counts.memories || 0) + (counts.timeline || 0);

    container.innerHTML = `
        <div class="memory-page">
            <div class="memory-dashboard">
                ${memoryStat('章节摘要', counts.summaries || 0, '每章剧情梗概与变化')}
                ${memoryStat('人物状态', counts.characters || 0, '当前位置、目标、能力')}
                ${memoryStat('剧情记忆', counts.memories || 0, '秘密、线索、关系、规则')}
                ${memoryStat('时间线', counts.timeline || 0, '故事内事件顺序')}
            </div>
            <div class="card memory-toolbar">
                <div>
                    <strong>长期记忆库</strong>
                    <p>用于长篇续写前回顾事实、状态、伏笔和时间线。</p>
                </div>
                <div class="memory-search">
                    <input id="memory-query" type="text" placeholder="检索人物、伏笔、道具、秘密、时间线..." onkeyup="if(event.key==='Enter')searchMemory()">
                    <button class="btn btn-primary" onclick="searchMemory()">检索</button>
                    <button class="btn btn-outline" onclick="loadMemory()">重置</button>
                </div>
            </div>
            <div id="memory-search-results"></div>
            ${total === 0 ? renderMemoryEmptyGuide() : `
                ${renderChapterSummaries(summaries)}
                ${renderCharacterStates(characters)}
                ${renderPlotMemories(memories)}
                ${renderTimeline(timeline)}
            `}
        </div>
    `;
}

function memoryStat(label, value, desc) {
    return `
        <div class="card memory-stat">
            <div class="memory-stat-value">${value}</div>
            <div class="memory-stat-label">${label}</div>
            <div class="memory-stat-desc">${desc}</div>
        </div>
    `;
}

function renderMemoryEmptyGuide() {
    return `
        <div class="card memory-empty-panel">
            <div class="memory-empty-title">还没有长期记忆</div>
            <p>可以先让 agent 为已有章节补录记忆。补录后，这里会显示章节摘要、人物当前状态、剧情事实和故事时间线。</p>
            <div class="memory-todo-grid">
                <div><strong>1. 章节摘要</strong><span>update_chapter_summary</span></div>
                <div><strong>2. 人物状态</strong><span>update_character_state</span></div>
                <div><strong>3. 剧情事实</strong><span>upsert_plot_memory</span></div>
                <div><strong>4. 时间线</strong><span>upsert_timeline_event</span></div>
            </div>
            <div class="memory-prompt-box">
请为 novel_id=当前小说ID 的第1-5章补录长期记忆。逐章读取正文，生成 chapter_summary，并同步人物当前状态、剧情事实记忆和时间线；不要改正文，完成后汇报补录结果。
            </div>
        </div>
    `;
}

async function searchMemory() {
    const query = document.getElementById('memory-query').value.trim();
    const target = document.getElementById('memory-search-results');
    if (!query) {
        target.innerHTML = '';
        return;
    }
    target.innerHTML = '<div class="memory-section"><div class="muted">正在检索...</div></div>';
    const res = await api('/novels/' + currentNovelId + '/memory/search?query=' + encodeURIComponent(query) + '&limit=30');
    if (res.code !== 0) {
        target.innerHTML = `<div class="memory-warning"><strong>检索失败</strong><p>${esc(res.message || '')}</p></div>`;
        return;
    }
    const results = res.data?.results || [];
    target.innerHTML = `
        <div class="memory-section">
            <div class="section-title">检索结果 <span>${results.length}</span></div>
            ${results.length ? results.map(r => `
                <div class="memory-item">
                    <div class="memory-item-head">
                        <strong>${esc(r.title)}</strong>
                        <span class="source-pill">${sourceName(r.source)}</span>
                    </div>
                    <p>${esc(r.snippet || r.content || '')}</p>
                    <div class="memory-meta">重要度 ${r.importance || 3} · ${formatDate(r.updated_at)}</div>
                </div>
            `).join('') : '<div class="empty-state compact"><p>没有命中长期记忆</p></div>'}
        </div>
    `;
}

function renderChapterSummaries(list) {
    return `
        <div class="memory-section">
            <div class="section-title">章节摘要 <span>${list.length}</span></div>
            ${list.length ? list.map(s => `
                <div class="memory-item">
                    <div class="memory-item-head">
                        <strong>${s.chapter_order}. ${esc(s.chapter_title)}</strong>
                        <button class="btn btn-outline btn-sm" onclick="navigate('/chapter/${s.chapter_id}')">查看章节</button>
                    </div>
                    <p>${esc(s.summary || '暂无摘要')}</p>
                    ${jsonChips(s.characters, '人物')}
                    ${jsonChips(s.locations, '地点')}
                    ${memoryBlock('关键事件', s.key_events)}
                    ${memoryBlock('剧情线', s.plot_threads)}
                    ${memoryBlock('伏笔变化', s.foreshadowing_changes)}
                    ${memoryBlock('人物变化', s.character_changes)}
                    <div class="memory-meta">${esc(s.timeline_position || '未记录时间位置')} · ${formatDate(s.updated_at)}</div>
                </div>
            `).join('') : emptyMemory('还没有章节摘要。可以让 agent 为前几章补录 update_chapter_summary。')}
        </div>
    `;
}

function renderCharacterStates(list) {
    return `
        <div class="memory-section">
            <div class="section-title">人物当前状态 <span>${list.length}</span></div>
            ${list.length ? `<div class="memory-grid">${list.map(ch => `
                <div class="memory-item">
                    <div class="memory-item-head">
                        <strong>${esc(ch.name)}${ch.alias ? ` (${esc(ch.alias)})` : ''}</strong>
                        ${ch.last_seen_chapter_title ? `<span class="source-pill">${esc(ch.last_seen_chapter_title)}</span>` : ''}
                    </div>
                    <p>${esc(ch.current_state || '暂无当前状态')}</p>
                    ${plainLine('位置', ch.location)}
                    ${plainLine('目标', ch.goal)}
                    ${plainLine('能力', ch.ability_state)}
                    ${plainLine('关系', ch.relationship_summary)}
                    ${plainLine('已知信息', ch.knowledge_state)}
                    <div class="memory-meta">${formatDate(ch.updated_at)}</div>
                </div>
            `).join('')}</div>` : emptyMemory('还没有人物当前状态。可以让 novel-lore 调用 update_character_state。')}
        </div>
    `;
}

function renderPlotMemories(list) {
    return `
        <div class="memory-section">
            <div class="section-title">剧情事实记忆 <span>${list.length}</span></div>
            ${list.length ? `<div class="memory-grid">${list.map(m => `
                <div class="memory-item">
                    <div class="memory-item-head">
                        <strong>${esc(m.title)}</strong>
                        <span class="source-pill">${memoryTypeName(m.memory_type)}</span>
                    </div>
                    <p>${esc(m.content || '')}</p>
                    ${jsonChips(m.tags, '标签')}
                    <div class="memory-meta">
                        重要度 ${m.importance} · ${memoryStatusName(m.status)}
                        ${m.chapter_title ? ' · ' + esc(m.chapter_title) : ''}
                        ${m.character_name ? ' · ' + esc(m.character_name) : ''}
                    </div>
                </div>
            `).join('')}</div>` : emptyMemory('还没有剧情事实记忆。可以让 novel-lore 调用 upsert_plot_memory。')}
        </div>
    `;
}

function renderTimeline(list) {
    return `
        <div class="memory-section">
            <div class="section-title">故事时间线 <span>${list.length}</span></div>
            ${list.length ? `<div class="timeline-list">${list.map(t => `
                <div class="timeline-row">
                    <div class="timeline-marker">${t.sequence_no || '-'}</div>
                    <div class="memory-item">
                        <div class="memory-item-head">
                            <strong>${esc(t.title)}</strong>
                            ${t.event_time ? `<span class="source-pill">${esc(t.event_time)}</span>` : ''}
                        </div>
                        <p>${esc(t.content || '')}</p>
                        <div class="memory-meta">重要度 ${t.importance}${t.chapter_title ? ' · ' + esc(t.chapter_title) : ''}</div>
                    </div>
                </div>
            `).join('')}</div>` : emptyMemory('还没有时间线事件。可以让 novel-lore 调用 upsert_timeline_event。')}
        </div>
    `;
}

function emptyMemory(text) {
    return `<div class="empty-state compact"><p>${esc(text)}</p></div>`;
}

function plainLine(label, value) {
    if (!value) return '';
    return `<div class="memory-line"><span>${label}</span>${esc(value)}</div>`;
}

function memoryBlock(label, raw) {
    const text = readableJSON(raw);
    if (!text) return '';
    return `<div class="memory-block"><strong>${label}</strong><div>${esc(text)}</div></div>`;
}

function jsonChips(raw, label) {
    const arr = parseJSONish(raw);
    if (!arr) return '';
    const values = Array.isArray(arr) ? arr : Object.entries(arr).map(([k, v]) => `${k}: ${typeof v === 'object' ? JSON.stringify(v) : v}`);
    if (!values.length) return '';
    return `<div class="memory-chips"><span>${label}</span>${values.map(v => `<em>${esc(String(v))}</em>`).join('')}</div>`;
}

function parseJSONish(raw) {
    if (!raw) return null;
    if (typeof raw === 'object') return raw;
    try { return JSON.parse(raw); } catch (_) { return null; }
}

function readableJSON(raw) {
    if (!raw) return '';
    const parsed = parseJSONish(raw);
    if (!parsed) return String(raw);
    if (Array.isArray(parsed)) {
        return parsed.map(item => typeof item === 'object' ? JSON.stringify(item) : String(item)).join('\n');
    }
    return Object.entries(parsed).map(([k, v]) => `${k}: ${typeof v === 'object' ? JSON.stringify(v) : v}`).join('\n');
}

function sourceName(s) {
    return { plot_memory: '剧情记忆', chapter_summary: '章节摘要', character_state: '人物状态', timeline: '时间线' }[s] || s;
}

function memoryTypeName(t) {
    return { fact: '事实', secret: '秘密', clue: '线索', rule: '规则', relationship: '关系', conflict: '冲突', plan: '计划' }[t] || t || '事实';
}

function memoryStatusName(s) {
    return { 0: '有效', 1: '已过期', 2: '存疑' }[s] || '有效';
}

function formatDate(value) {
    if (!value) return '';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return date.toLocaleString();
}

// --- Modal ---
function openModal(html) {
    document.getElementById('modal').innerHTML = html;
    document.getElementById('modal').classList.remove('hidden');
    document.getElementById('modal-overlay').classList.remove('hidden');
}

function closeModal() {
    document.getElementById('modal').classList.add('hidden');
    document.getElementById('modal-overlay').classList.add('hidden');
}

// --- Utils ---
function esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

// --- Version History (read-only) ---
function showVersionHistory(entityType, entityId) {
    api('/versions/' + entityType + '/' + entityId).then(res => {
        const versions = res.data || [];
        const typeLabel = {novel:'小说',chapter:'章节',character:'人物',worldview:'世界观',foreshadowing:'伏笔'}[entityType] || entityType;
        let html = `<h3 style="margin-bottom:16px">📋 ${typeLabel}版本历史</h3>`;
        if (!versions.length) {
            html += '<p style="color:var(--text-secondary)">暂无历史版本</p>';
        } else {
            html += `<div class="version-list">`;
            versions.forEach(v => {
                const snapshot = typeof v.snapshot === 'string' ? JSON.parse(v.snapshot) : v.snapshot;
                const title = snapshot.title || snapshot.name || '无标题';
                html += `
                    <div class="version-item">
                        <div style="display:flex;justify-content:space-between;align-items:center">
                            <div>
                                <span class="version-badge">v${v.version}</span>
                                <strong>${esc(title)}</strong>
                                <span style="font-size:12px;color:var(--text-secondary);margin-left:8px">${v.change_summary || ''}</span>
                            </div>
                            <div class="btn-group">
                                <button class="btn btn-outline btn-sm" onclick="previewVersion(${v.id})">预览</button>
                            </div>
                        </div>
                        <div style="font-size:12px;color:var(--text-secondary);margin-top:4px">${new Date(v.created_at).toLocaleString()}</div>
                    </div>
                `;
            });
            html += '</div>';
        }
        html += `<div style="margin-top:16px;text-align:right"><button class="btn btn-outline" onclick="closeModal()">关闭</button></div>`;
        openModal(html);
    });
}

async function previewVersion(versionId) {
    const res = await api('/versions/detail/' + versionId);
    const v = res.data;
    if (!v) return;
    const snapshot = typeof v.snapshot === 'string' ? JSON.parse(v.snapshot) : v.snapshot;
    let html = `<h3 style="margin-bottom:16px">📋 版本 v${v.version} 预览</h3>`;
    html += `<div class="version-preview">`;
    for (const [key, value] of Object.entries(snapshot)) {
        if (key === 'id' || key === 'novel_id' || key === 'created_at' || key === 'updated_at') continue;
        const label = {
            title: '标题', name: '姓名', author: '作者', description: '描述', cover_url: '封面URL',
            status: '状态', content: '内容', word_count: '字数', chapter_order: '章节序号',
            alias: '别名', avatar_url: '头像URL', gender: '性别', age: '年龄',
            personality: '性格', background: '背景', category: '分类', sort_order: '排序',
            character_order: '排序', planted_chapter_id: '埋设章节ID', resolved_chapter_id: '回收章节ID',
            importance: '重要程度'
        }[key] || key;
        if (value === null || value === undefined || value === '') continue;
        html += `<div class="form-group"><label>${label}</label>`;
        if (typeof value === 'string' && value.length > 80) {
            html += `<div class="version-preview-content">${esc(String(value))}</div>`;
        } else {
            html += `<div>${esc(String(value))}</div>`;
        }
        html += `</div>`;
    }
    html += `</div>`;
    html += `<div style="margin-top:16px;text-align:right">
        <button class="btn btn-outline" onclick="showVersionHistory('${v.entity_type}', ${v.entity_id})">返回列表</button>
    </div>`;
    openModal(html);
}

// Init
document.addEventListener('DOMContentLoaded', handleRoute);
