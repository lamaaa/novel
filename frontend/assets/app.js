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
}

function tabName(t) {
    return { chapters: '目录', characters: '人物', worldviews: '世界观', foreshadowings: '伏笔' }[t] || t;
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
