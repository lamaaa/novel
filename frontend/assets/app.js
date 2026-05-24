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
    } else if (path === '/novel/create') {
        renderNovelForm();
    } else if (path.match(/^\/novel\/\d+\/edit$/)) {
        const id = path.split('/')[2];
        renderNovelForm(id);
    } else if (path.match(/^\/novel\/\d+$/)) {
        const id = path.split('/')[2];
        renderNovelDetail(id);
    } else if (path.match(/^\/chapter\/\d+$/)) {
        const id = path.split('/')[2];
        renderChapterReading(id);
    } else if (path.match(/^\/novel\/\d+\/chapter\/create$/)) {
        const novelId = path.split('/')[2];
        renderChapterForm(novelId);
    } else if (path.match(/^\/chapter\/\d+\/edit$/)) {
        const id = path.split('/')[2];
        renderChapterForm(null, id);
    } else if (path.match(/^\/novel\/\d+\/character\/create$/)) {
        const novelId = path.split('/')[2];
        renderCharacterForm(novelId);
    } else if (path.match(/^\/character\/\d+\/edit$/)) {
        const id = path.split('/')[2];
        renderCharacterForm(null, id);
    } else if (path.match(/^\/novel\/\d+\/worldview\/create$/)) {
        const novelId = path.split('/')[2];
        renderWorldviewForm(novelId);
    } else if (path.match(/^\/worldview\/\d+\/edit$/)) {
        const id = path.split('/')[2];
        renderWorldviewForm(null, id);
    } else if (path.match(/^\/novel\/\d+\/foreshadowing\/create$/)) {
        const novelId = path.split('/')[2];
        renderForeshadowingForm(novelId);
    } else if (path.match(/^\/foreshadowing\/\d+\/edit$/)) {
        const id = path.split('/')[2];
        renderForeshadowingForm(null, id);
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
            <button class="btn btn-primary" onclick="navigate('/novel/create')">✏️ 写新书</button>
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
        container.innerHTML = '<div class="empty-state"><div class="icon">📖</div><p>还没有小说，开始创作吧！</p></div>';
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

// --- Novel Form ---
async function renderNovelForm(id) {
    const app = document.getElementById('app');
    let novel = { title: '', author: '', description: '', cover_url: '', status: 0 };
    if (id) {
        const res = await api('/novels/' + id);
        novel = res.data || novel;
    }
    app.innerHTML = `
        <div class="back-link" onclick="navigate('/')">← 返回书架</div>
        <div class="card">
            <h2 style="margin-bottom:20px">${id ? '编辑小说' : '创建小说'}</h2>
            <form onsubmit="saveNovel(event, ${id || 0})">
                <div class="form-group"><label>标题 *</label><input id="f-title" value="${esc(novel.title)}" required></div>
                <div class="form-group"><label>作者</label><input id="f-author" value="${esc(novel.author)}"></div>
                <div class="form-group"><label>封面URL</label><input id="f-cover" value="${esc(novel.cover_url)}"></div>
                <div class="form-group"><label>简介</label><textarea id="f-desc">${esc(novel.description)}</textarea></div>
                <div class="form-group"><label>状态</label><select id="f-status">
                    <option value="0" ${novel.status===0?'selected':''}>连载中</option>
                    <option value="1" ${novel.status===1?'selected':''}>已完结</option>
                </select></div>
                <div class="btn-group"><button type="submit" class="btn btn-primary">保存</button>${id ? `<button type="button" class="btn btn-outline" onclick="showVersionHistory('novel', ${id})">📋 版本历史</button>` : ''}<button type="button" class="btn btn-outline" onclick="history.back()">取消</button></div>
            </form>
        </div>
    `;
}

async function saveNovel(e, id) {
    e.preventDefault();
    const data = {
        title: document.getElementById('f-title').value,
        author: document.getElementById('f-author').value,
        cover_url: document.getElementById('f-cover').value,
        description: document.getElementById('f-desc').value,
        status: parseInt(document.getElementById('f-status').value),
    };
    if (id) {
        await api('/novels/' + id, { method: 'PUT', body: JSON.stringify(data) });
    } else {
        await api('/novels', { method: 'POST', body: JSON.stringify(data) });
    }
    navigate('/');
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
                <div class="btn-group">
                    <button class="btn btn-primary btn-sm" onclick="navigate('/novel/${id}/chapter/create')">+ 添加章节</button>
                    <button class="btn btn-outline btn-sm" onclick="navigate('/novel/${id}/edit')">编辑</button>
                    <button class="btn btn-danger btn-sm" onclick="deleteNovel(${id})">删除</button>
                </div>
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
        container.innerHTML = `<div class="empty-state"><div class="icon">📝</div><p>还没有章节</p>
            <button class="btn btn-primary" style="margin-top:12px" onclick="navigate('/novel/${currentNovelId}/chapter/create')">添加第一章</button></div>`;
        return;
    }
    container.innerHTML = `
        <div style="margin-bottom:12px"><button class="btn btn-primary btn-sm" onclick="navigate('/novel/${currentNovelId}/chapter/create')">+ 添加章节</button></div>
        <ul class="chapter-list card" style="padding:0">
        ${chapters.map(ch => `
            <li onclick="navigate('/chapter/${ch.id}')">
                <div><span class="title">${ch.chapter_order}. ${esc(ch.title)}</span> ${ch.status===0?'<span class="draft-tag">草稿</span>':''}</div>
                <div class="meta">${ch.word_count}字 <button class="btn btn-outline btn-sm" onclick="event.stopPropagation();navigate('/chapter/${ch.id}/edit')">编辑</button></div>
            </li>
        `).join('')}
        </ul>
    `;
}

// --- Chapter Form ---
async function renderChapterForm(novelId, id) {
    const app = document.getElementById('app');
    let chapter = { title: '', content: '', chapter_order: 0, status: 0 };
    let currentNovel = novelId;

    if (id) {
        const res = await api('/chapters/' + id);
        chapter = res.data?.chapter || chapter;
        currentNovel = chapter.novel_id;
    }

    // get chapters for order reference
    const chaptersRes = await api('/novels/' + currentNovel + '/chapters');
    const chapters = chaptersRes.data || [];

    app.innerHTML = `
        <div class="back-link" onclick="navigate('/novel/${currentNovel}')">← 返回小说</div>
        <div class="card">
            <h2 style="margin-bottom:20px">${id ? '编辑章节' : '添加章节'}</h2>
            <form onsubmit="saveChapter(event, ${currentNovel}, ${id || 0})">
                <div class="form-group"><label>章节标题 *</label><input id="f-ch-title" value="${esc(chapter.title)}" required></div>
                <div class="form-group"><label>章节排序 (0=自动)</label><input type="number" id="f-ch-order" value="${chapter.chapter_order}"></div>
                <div class="form-group"><label>内容</label><textarea id="f-ch-content" style="min-height:400px">${esc(chapter.content)}</textarea></div>
                <div class="form-group"><label>状态</label><select id="f-ch-status">
                    <option value="0" ${chapter.status===0?'selected':''}>草稿</option>
                    <option value="1" ${chapter.status===1?'selected':''}>已发布</option>
                </select></div>
                <div class="btn-group"><button type="submit" class="btn btn-primary">保存</button>${id ? `<button type="button" class="btn btn-outline" onclick="showVersionHistory('chapter', ${id})">📋 版本历史</button>` : ''}<button type="button" class="btn btn-outline" onclick="history.back()">取消</button></div>
            </form>
        </div>
    `;
}

async function saveChapter(e, novelId, id) {
    e.preventDefault();
    const data = {
        title: document.getElementById('f-ch-title').value,
        content: document.getElementById('f-ch-content').value,
        chapter_order: parseInt(document.getElementById('f-ch-order').value),
        status: parseInt(document.getElementById('f-ch-status').value),
    };
    if (id) {
        await api('/chapters/' + id, { method: 'PUT', body: JSON.stringify(data) });
        navigate('/chapter/' + id);
    } else {
        await api('/novels/' + novelId + '/chapters', { method: 'POST', body: JSON.stringify(data) });
        navigate('/novel/' + novelId);
    }
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
                <button class="btn btn-outline" onclick="navigate('/chapter/${ch.id}/edit')">编辑</button>
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
        container.innerHTML = `<div class="empty-state"><div class="icon">👤</div><p>还没有人物</p>
            <button class="btn btn-primary" style="margin-top:12px" onclick="navigate('/novel/${currentNovelId}/character/create')">添加人物</button></div>`;
        return;
    }
    container.innerHTML = `
        <div style="margin-bottom:12px"><button class="btn btn-primary btn-sm" onclick="navigate('/novel/${currentNovelId}/character/create')">+ 添加人物</button></div>
        <div class="character-grid">
        ${characters.map(ch => `
            <div class="card character-card card-clickable" onclick="showCharacterDetail(${ch.id})">
                <div class="avatar">${ch.avatar_url ? `<img src="${ch.avatar_url}" style="width:100%;height:100%;border-radius:50%;object-fit:cover">` : '👤'}</div>
                <div class="info">
                    <h4>${esc(ch.name)}${ch.alias ? ` (${esc(ch.alias)})` : ''}</h4>
                    <p>${genderText(ch.gender)}${ch.age ? ' · '+ch.age+'岁' : ''}</p>
                    <p style="margin-top:4px">${esc((ch.description||'').substring(0,60))}${(ch.description||'').length>60?'...':''}</p>
                </div>
            </div>
        `).join('')}
        </div>
    `;
}

function genderText(g) { return {0:'未知',1:'男',2:'女'}[g]||'未知'; }

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
        <div class="btn-group" style="margin-top:16px">
            <button class="btn btn-outline btn-sm" onclick="closeModal();navigate('/character/${ch.id}/edit')">编辑</button>
            <button class="btn btn-danger btn-sm" onclick="closeModal();deleteCharacter(${ch.id})">删除</button>
        </div>
    `);
}

// --- Character Form ---
async function renderCharacterForm(novelId, id) {
    const app = document.getElementById('app');
    let ch = { name:'', alias:'', avatar_url:'', gender:0, age:'', description:'', personality:'', background:'', character_order:0 };
    let currentNovel = novelId;

    if (id) {
        const res = await api('/characters/' + id);
        ch = res.data || ch;
        currentNovel = ch.novel_id;
    }

    app.innerHTML = `
        <div class="back-link" onclick="navigate('/novel/${currentNovel}')">← 返回小说</div>
        <div class="card">
            <h2 style="margin-bottom:20px">${id ? '编辑人物' : '添加人物'}</h2>
            <form onsubmit="saveCharacter(event, ${currentNovel}, ${id || 0})">
                <div class="form-group"><label>姓名 *</label><input id="f-char-name" value="${esc(ch.name)}" required></div>
                <div class="form-group"><label>别名/外号</label><input id="f-char-alias" value="${esc(ch.alias)}"></div>
                <div class="form-group"><label>头像URL</label><input id="f-char-avatar" value="${esc(ch.avatar_url)}"></div>
                <div class="form-group"><label>性别</label><select id="f-char-gender">
                    <option value="0" ${ch.gender===0?'selected':''}>未知</option>
                    <option value="1" ${ch.gender===1?'selected':''}>男</option>
                    <option value="2" ${ch.gender===2?'selected':''}>女</option>
                </select></div>
                <div class="form-group"><label>年龄</label><input id="f-char-age" value="${esc(ch.age)}"></div>
                <div class="form-group"><label>描述</label><textarea id="f-char-desc">${esc(ch.description)}</textarea></div>
                <div class="form-group"><label>性格特点</label><textarea id="f-char-personality">${esc(ch.personality)}</textarea></div>
                <div class="form-group"><label>背景故事</label><textarea id="f-char-background">${esc(ch.background)}</textarea></div>
                <div class="btn-group"><button type="submit" class="btn btn-primary">保存</button>${id ? `<button type="button" class="btn btn-outline" onclick="showVersionHistory('character', ${id})">📋 版本历史</button>` : ''}<button type="button" class="btn btn-outline" onclick="history.back()">取消</button></div>
            </form>
        </div>
    `;
}

async function saveCharacter(e, novelId, id) {
    e.preventDefault();
    const data = {
        name: document.getElementById('f-char-name').value,
        alias: document.getElementById('f-char-alias').value,
        avatar_url: document.getElementById('f-char-avatar').value,
        gender: parseInt(document.getElementById('f-char-gender').value),
        age: document.getElementById('f-char-age').value,
        description: document.getElementById('f-char-desc').value,
        personality: document.getElementById('f-char-personality').value,
        background: document.getElementById('f-char-background').value,
    };
    if (id) {
        await api('/characters/' + id, { method: 'PUT', body: JSON.stringify(data) });
    } else {
        await api('/novels/' + novelId + '/characters', { method: 'POST', body: JSON.stringify(data) });
    }
    navigate('/novel/' + novelId);
}

async function deleteCharacter(id) {
    if (!confirm('确定删除此人物？')) return;
    await api('/characters/' + id, { method: 'DELETE' });
    navigate('/novel/' + currentNovelId);
}

// --- Worldviews ---
async function loadWorldviews() {
    const res = await api('/novels/' + currentNovelId + '/worldviews');
    const data = res.data || {};
    const worldviews = data.list || [];
    const categories = data.categories || [];
    const container = document.getElementById('tab-content');

    if (!worldviews.length) {
        container.innerHTML = `<div class="empty-state"><div class="icon">🌍</div><p>还没有世界观设定</p>
            <button class="btn btn-primary" style="margin-top:12px" onclick="navigate('/novel/${currentNovelId}/worldview/create')">添加设定</button></div>`;
        return;
    }

    let html = `<div style="margin-bottom:12px"><button class="btn btn-primary btn-sm" onclick="navigate('/novel/${currentNovelId}/worldview/create')">+ 添加设定</button></div>`;

    categories.forEach(cat => {
        const items = worldviews.filter(w => w.category === cat);
        html += `<div class="worldview-group"><h3>📍 ${esc(cat)}</h3>`;
        items.forEach(w => {
            html += `<div class="card" style="margin-bottom:8px">
                <div style="display:flex;justify-content:space-between;align-items:center">
                    <strong>${esc(w.title)}</strong>
                    <div class="btn-group">
                        <button class="btn btn-outline btn-sm" onclick="navigate('/worldview/${w.id}/edit')">编辑</button>
                        <button class="btn btn-danger btn-sm" onclick="deleteWorldview(${w.id})">删除</button>
                    </div>
                </div>
                <p style="margin-top:8px;color:var(--text-secondary);white-space:pre-wrap">${esc(w.content)}</p>
            </div>`;
        });
        html += '</div>';
    });

    container.innerHTML = html;
}

// --- Worldview Form ---
async function renderWorldviewForm(novelId, id) {
    const app = document.getElementById('app');
    let w = { category: '其他', title: '', content: '', sort_order: 0 };
    let currentNovel = novelId;

    if (id) {
        const res = await api('/worldviews/' + id);
        w = res.data || w;
        currentNovel = w.novel_id;
    }

    app.innerHTML = `
        <div class="back-link" onclick="navigate('/novel/${currentNovel}')">← 返回小说</div>
        <div class="card">
            <h2 style="margin-bottom:20px">${id ? '编辑世界观设定' : '添加世界观设定'}</h2>
            <form onsubmit="saveWorldview(event, ${currentNovel}, ${id || 0})">
                <div class="form-group"><label>分类 *</label><input id="f-wv-category" value="${esc(w.category)}" placeholder="如：地理、历史、种族、魔法体系、势力..." required></div>
                <div class="form-group"><label>标题 *</label><input id="f-wv-title" value="${esc(w.title)}" required></div>
                <div class="form-group"><label>内容</label><textarea id="f-wv-content" style="min-height:200px">${esc(w.content)}</textarea></div>
                <div class="btn-group"><button type="submit" class="btn btn-primary">保存</button>${id ? `<button type="button" class="btn btn-outline" onclick="showVersionHistory('worldview', ${id})">📋 版本历史</button>` : ''}<button type="button" class="btn btn-outline" onclick="history.back()">取消</button></div>
            </form>
        </div>
    `;
}

async function saveWorldview(e, novelId, id) {
    e.preventDefault();
    const data = {
        category: document.getElementById('f-wv-category').value,
        title: document.getElementById('f-wv-title').value,
        content: document.getElementById('f-wv-content').value,
    };
    if (id) {
        await api('/worldviews/' + id, { method: 'PUT', body: JSON.stringify(data) });
    } else {
        await api('/novels/' + novelId + '/worldviews', { method: 'POST', body: JSON.stringify(data) });
    }
    navigate('/novel/' + novelId);
}

async function deleteWorldview(id) {
    if (!confirm('确定删除此设定？')) return;
    await api('/worldviews/' + id, { method: 'DELETE' });
    switchTab('worldviews');
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
        container.innerHTML = `<div class="empty-state"><div class="icon">🎯</div><p>还没有伏笔</p>
            <button class="btn btn-primary" style="margin-top:12px" onclick="navigate('/novel/${currentNovelId}/foreshadowing/create')">添加伏笔</button></div>`;
        return;
    }

    const statusMap = {0:'已埋设',1:'已回收',2:'已放弃'};
    container.innerHTML = `
        <div style="margin-bottom:12px"><button class="btn btn-primary btn-sm" onclick="navigate('/novel/${currentNovelId}/foreshadowing/create')">+ 添加伏笔</button></div>
        ${list.map(f => `
            <div class="card">
                <div style="display:flex;justify-content:space-between;align-items:flex-start">
                    <div>
                        <strong>${esc(f.title)}</strong>
                        <span class="foreshadowing-status fs-status-${f.status}">${statusMap[f.status]}</span>
                        <div class="importance-bar">${[1,2,3,4,5].map(i=>`<div class="dot ${i<=f.importance?'filled':''}"></div>`).join('')}</div>
                    </div>
                    <div class="btn-group">
                        <button class="btn btn-outline btn-sm" onclick="navigate('/foreshadowing/${f.id}/edit')">编辑</button>
                        <button class="btn btn-danger btn-sm" onclick="deleteForeshadowing(${f.id})">删除</button>
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

// --- Foreshadowing Form ---
async function renderForeshadowingForm(novelId, id) {
    const app = document.getElementById('app');
    let f = { title:'', description:'', planted_chapter_id:null, resolved_chapter_id:null, status:0, importance:3 };
    let currentNovel = novelId;

    if (id) {
        const res = await api('/foreshadowings/' + id);
        f = res.data || f;
        currentNovel = f.novel_id;
    }

    const chaptersRes = await api('/novels/' + currentNovel + '/chapters');
    const chapters = chaptersRes.data || [];

    app.innerHTML = `
        <div class="back-link" onclick="navigate('/novel/${currentNovel}')">← 返回小说</div>
        <div class="card">
            <h2 style="margin-bottom:20px">${id ? '编辑伏笔' : '添加伏笔'}</h2>
            <form onsubmit="saveForeshadowing(event, ${currentNovel}, ${id || 0})">
                <div class="form-group"><label>伏笔标题 *</label><input id="f-fs-title" value="${esc(f.title)}" required></div>
                <div class="form-group"><label>描述</label><textarea id="f-fs-desc">${esc(f.description)}</textarea></div>
                <div class="form-group"><label>埋设章节</label><select id="f-fs-planted">
                    <option value="">未指定</option>
                    ${chapters.map(ch=>`<option value="${ch.id}" ${f.planted_chapter_id==ch.id?'selected':''}>${ch.chapter_order}. ${esc(ch.title)}</option>`).join('')}
                </select></div>
                <div class="form-group"><label>回收章节</label><select id="f-fs-resolved">
                    <option value="">未指定</option>
                    ${chapters.map(ch=>`<option value="${ch.id}" ${f.resolved_chapter_id==ch.id?'selected':''}>${ch.chapter_order}. ${esc(ch.title)}</option>`).join('')}
                </select></div>
                <div class="form-group"><label>状态</label><select id="f-fs-status">
                    <option value="0" ${f.status===0?'selected':''}>已埋设</option>
                    <option value="1" ${f.status===1?'selected':''}>已回收</option>
                    <option value="2" ${f.status===2?'selected':''}>已放弃</option>
                </select></div>
                <div class="form-group"><label>重要程度 (1-5)</label><input type="number" id="f-fs-importance" min="1" max="5" value="${f.importance}"></div>
                <div class="btn-group"><button type="submit" class="btn btn-primary">保存</button>${id ? `<button type="button" class="btn btn-outline" onclick="showVersionHistory('foreshadowing', ${id})">📋 版本历史</button>` : ''}<button type="button" class="btn btn-outline" onclick="history.back()">取消</button></div>
            </form>
        </div>
    `;
}

async function saveForeshadowing(e, novelId, id) {
    e.preventDefault();
    const planted = document.getElementById('f-fs-planted').value;
    const resolved = document.getElementById('f-fs-resolved').value;
    const data = {
        title: document.getElementById('f-fs-title').value,
        description: document.getElementById('f-fs-desc').value,
        planted_chapter_id: planted ? parseInt(planted) : null,
        resolved_chapter_id: resolved ? parseInt(resolved) : null,
        status: parseInt(document.getElementById('f-fs-status').value),
        importance: parseInt(document.getElementById('f-fs-importance').value),
    };
    if (id) {
        await api('/foreshadowings/' + id, { method: 'PUT', body: JSON.stringify(data) });
    } else {
        await api('/novels/' + novelId + '/foreshadowings', { method: 'POST', body: JSON.stringify(data) });
    }
    navigate('/novel/' + novelId);
}

async function deleteForeshadowing(id) {
    if (!confirm('确定删除此伏笔？')) return;
    await api('/foreshadowings/' + id, { method: 'DELETE' });
    switchTab('foreshadowings');
}

// --- Delete Novel ---
async function deleteNovel(id) {
    if (!confirm('确定删除此小说？所有章节、人物、世界观和伏笔数据将被删除！')) return;
    await api('/novels/' + id, { method: 'DELETE' });
    navigate('/');
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

// --- Version History ---
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
                                <button class="btn btn-primary btn-sm" onclick="rollbackVersion(${v.id}, '${entityType}', ${entityId})">回退</button>
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
    html += `<div style="margin-top:16px" class="btn-group">
        <button class="btn btn-primary" onclick="rollbackVersion(${v.id}, '${v.entity_type}', ${v.entity_id})">回退到此版本</button>
        <button class="btn btn-outline" onclick="showVersionHistory('${v.entity_type}', ${v.entity_id})">返回列表</button>
    </div>`;
    openModal(html);
}

async function rollbackVersion(versionId, entityType, entityId) {
    if (!confirm('确定回退到此版本？当前状态将自动保存为新版本。')) return;
    const res = await api('/versions/' + versionId + '/rollback', {
        method: 'POST',
        body: JSON.stringify({ change_summary: '手动回退' })
    });
    if (res.code === 0) {
        alert('回退成功！');
        closeModal();
        // 刷新当前页面
        handleRoute();
    } else {
        alert('回退失败：' + res.message);
    }
}

// Init
document.addEventListener('DOMContentLoaded', handleRoute);
