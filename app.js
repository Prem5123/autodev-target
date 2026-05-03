// Scribe - Clinical Notes App

const API_BASE = '/api';

// State
let isAuthenticated = false;
let notes = [];
let currentEditId = null;

// DOM Elements
const loginView = document.getElementById('login-view');
const mainView = document.getElementById('main-view');
const loginForm = document.getElementById('login-form');
const loginError = document.getElementById('login-error');
const notesList = document.getElementById('notes-list');
const noteEditor = document.getElementById('note-editor');
const noteForm = document.getElementById('note-form');
const noteIdInput = document.getElementById('note-id');
const noteTitleInput = document.getElementById('note-title');
const noteTemplateSelect = document.getElementById('note-template');
const noteContentTextarea = document.getElementById('note-content');
const soapFields = document.getElementById('soap-fields');
const simpleField = document.getElementById('simple-field');
const newNoteBtn = document.getElementById('new-note-btn');
const logoutBtn = document.getElementById('logout-btn');
const closeEditorBtn = document.getElementById('close-editor-btn');
const cancelBtn = document.getElementById('cancel-btn');
const deleteNoteBtn = document.getElementById('delete-note-btn');
const editorTitle = document.getElementById('editor-title');

// SOAP field elements
const soapSubjective = document.getElementById('soap-subjective');
const soapObjective = document.getElementById('soap-objective');
const soapAssessment = document.getElementById('soap-assessment');
const soapPlan = document.getElementById('soap-plan');

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    checkAuth();
    setupEventListeners();
});

// Check authentication status
async function checkAuth() {
    try {
        const response = await fetch(`${API_BASE}/notes`);
        if (response.ok) {
            isAuthenticated = true;
            showMainView();
            loadNotes();
        } else {
            showLoginView();
        }
    } catch (err) {
        showLoginView();
    }
}

// Setup event listeners
function setupEventListeners() {
    loginForm.addEventListener('submit', handleLogin);
    logoutBtn.addEventListener('click', handleLogout);
    newNoteBtn.addEventListener('click', openNewNote);
    closeEditorBtn.addEventListener('click', closeEditor);
    cancelBtn.addEventListener('click', closeEditor);
    noteForm.addEventListener('submit', handleNoteSave);
    deleteNoteBtn.addEventListener('click', handleNoteDelete);
    noteTemplateSelect.addEventListener('change', toggleTemplateFields);
}

// Auth handlers
async function handleLogin(e) {
    e.preventDefault();
    loginError.classList.add('hidden');

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch(`${API_BASE}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            isAuthenticated = true;
            showMainView();
            loadNotes();
        } else {
            loginError.textContent = 'Invalid username or password';
            loginError.classList.remove('hidden');
        }
    } catch (err) {
        loginError.textContent = 'Connection error. Please try again.';
        loginError.classList.remove('hidden');
    }
}

async function handleLogout() {
    try {
        await fetch(`${API_BASE}/auth/logout`, { method: 'POST' });
    } catch (err) {
        // Ignore logout errors
    }
    isAuthenticated = false;
    showLoginView();
}

// View management
function showLoginView() {
    loginView.classList.remove('hidden');
    mainView.classList.add('hidden');
}

function showMainView() {
    loginView.classList.add('hidden');
    mainView.classList.remove('hidden');
}

// Notes API
async function loadNotes() {
    try {
        const response = await fetch(`${API_BASE}/notes`);
        if (response.ok) {
            notes = await response.json();
            renderNotes();
        }
    } catch (err) {
        console.error('Failed to load notes:', err);
    }
}

async function createNote(title, content, template) {
    const response = await fetch(`${API_BASE}/notes`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, template })
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
    }

    return response.json();
}

async function updateNote(id, title, content, template) {
    const response = await fetch(`${API_BASE}/notes/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, content, template })
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
    }

    return response.json();
}

async function deleteNote(id) {
    const response = await fetch(`${API_BASE}/notes/${id}`, {
        method: 'DELETE'
    });

    if (!response.ok) {
        const error = await response.text();
        throw new Error(error);
    }
}

// Render notes list
function renderNotes() {
    if (notes.length === 0) {
        notesList.innerHTML = '<p class="empty-state">No notes yet. Create your first note.</p>';
        return;
    }

    notesList.innerHTML = notes.map(note => `
        <div class="note-card" data-id="${note.id}">
            <div class="note-card-header">
                <span class="note-card-title">${escapeHtml(note.title)}</span>
                <span class="note-card-template ${note.template.toLowerCase()}">${note.template}</span>
            </div>
            <p class="note-card-preview">${escapeHtml(note.content)}</p>
            <p class="note-card-date">${formatDate(note.created_at)}</p>
        </div>
    `).join('');

    // Add click handlers
    notesList.querySelectorAll('.note-card').forEach(card => {
        card.addEventListener('click', () => openNote(card.dataset.id));
    });
}

// Note editor
function openNewNote() {
    currentEditId = null;
    editorTitle.textContent = 'New Note';
    noteForm.reset();
    noteIdInput.value = '';
    deleteNoteBtn.classList.add('hidden');
    noteEditor.classList.remove('hidden');
    toggleTemplateFields();
    noteTitleInput.focus();
}

function openNote(id) {
    const note = notes.find(n => n.id === id);
    if (!note) return;

    currentEditId = id;
    editorTitle.textContent = 'Edit Note';
    noteIdInput.value = note.id;
    noteTitleInput.value = note.title;
    noteTemplateSelect.value = note.template;

    if (note.template === 'SOAP') {
        // Parse SOAP content from note.content
        const soapParts = parseSoapContent(note.content);
        soapSubjective.value = soapParts.subjective;
        soapObjective.value = soapParts.objective;
        soapAssessment.value = soapParts.assessment;
        soapPlan.value = soapParts.plan;
    } else {
        noteContentTextarea.value = note.content;
    }

    deleteNoteBtn.classList.remove('hidden');
    noteEditor.classList.remove('hidden');
    toggleTemplateFields();
    noteTitleInput.focus();
}

function closeEditor() {
    noteEditor.classList.add('hidden');
    currentEditId = null;
}

function toggleTemplateFields() {
    const template = noteTemplateSelect.value;
    if (template === 'SOAP') {
        soapFields.classList.remove('hidden');
        simpleField.classList.add('hidden');
    } else {
        soapFields.classList.add('hidden');
        simpleField.classList.remove('hidden');
    }
}

// Parse SOAP content into parts
function parseSoapContent(content) {
    const parts = { subjective: '', objective: '', assessment: '', plan: '' };
    if (!content) return parts;

    // Simple parsing: look for S:, O:, A:, P: markers
    const lines = content.split('\n');
    let currentSection = null;

    lines.forEach(line => {
        const trimmed = line.trim();
        if (trimmed.startsWith('S:') || trimmed.startsWith('S -')) {
            currentSection = 'subjective';
            parts.subjective = trimmed.substring(2).trim();
        } else if (trimmed.startsWith('O:') || trimmed.startsWith('O -')) {
            currentSection = 'objective';
            parts.objective = trimmed.substring(2).trim();
        } else if (trimmed.startsWith('A:') || trimmed.startsWith('A -')) {
            currentSection = 'assessment';
            parts.assessment = trimmed.substring(2).trim();
        } else if (trimmed.startsWith('P:') || trimmed.startsWith('P -')) {
            currentSection = 'plan';
            parts.plan = trimmed.substring(2).trim();
        } else if (currentSection) {
            parts[currentSection] += '\n' + trimmed;
        }
    });

    return parts;
}

// Build SOAP content from parts
function buildSoapContent() {
    const parts = [];
    if (soapSubjective.value.trim()) parts.push(`S: ${soapSubjective.value.trim()}`);
    if (soapObjective.value.trim()) parts.push(`O: ${soapObjective.value.trim()}`);
    if (soapAssessment.value.trim()) parts.push(`A: ${soapAssessment.value.trim()}`);
    if (soapPlan.value.trim()) parts.push(`P: ${soapPlan.value.trim()}`);
    return parts.join('\n\n');
}

// Form handlers
async function handleNoteSave(e) {
    e.preventDefault();

    const title = noteTitleInput.value.trim();
    const template = noteTemplateSelect.value;
    const content = template === 'SOAP' ? buildSoapContent() : noteContentTextarea.value.trim();

    if (!title && !content) {
        alert('Please provide a title or content');
        return;
    }

    try {
        if (currentEditId) {
            await updateNote(currentEditId, title, content, template);
        } else {
            await createNote(title, content, template);
        }
        closeEditor();
        await loadNotes();
    } catch (err) {
        alert('Failed to save note: ' + err.message);
    }
}

async function handleNoteDelete() {
    if (!currentEditId) return;
    if (!confirm('Are you sure you want to delete this note?')) return;

    try {
        await deleteNote(currentEditId);
        closeEditor();
        await loadNotes();
    } catch (err) {
        alert('Failed to delete note: ' + err.message);
    }
}

// Utilities
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatDate(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}
