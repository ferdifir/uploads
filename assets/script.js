const API_URL = window.location.origin + "/api";
const API_KEY_STORAGE = 'RahasiaAPIKey123';

// Pagination variables
let currentPage = 1;
let itemsPerPage = 10;
let allFiles = [];

document.addEventListener('DOMContentLoaded', () => {
    // Cek apakah API Key sudah ada di Local Storage
    if (localStorage.getItem(API_KEY_STORAGE)) {
        showMainContainer();
        listFiles();
    } else {
        showLoginContainer();
    }
});

// --- UI MANAGEMENT ---

function showLoginContainer() {
    document.getElementById('login-container').style.display = 'block';
    document.getElementById('main-container').style.display = 'none';
}

function showMainContainer() {
    document.getElementById('login-container').style.display = 'none';
    document.getElementById('main-container').style.display = 'block';
}

function getAPIKey() {
    return localStorage.getItem(API_KEY_STORAGE);
}

// --- MODAL MANAGEMENT ---

function openUploadModal() {
    document.getElementById('upload-modal').classList.remove('hidden');
    document.getElementById('file-input').value = ''; // Clear any previously selected file
    document.getElementById('upload-status').textContent = '';
}

function closeUploadModal() {
    document.getElementById('upload-modal').classList.add('hidden');
}

// --- AUTHENTIKASI ---

async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const msgElement = document.getElementById('login-message');

    msgElement.textContent = '';

    try {
        const response = await fetch(`${API_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (response.ok) {
            localStorage.setItem(API_KEY_STORAGE, data.api_key);
            showMainContainer();
            listFiles();
        } else {
            msgElement.textContent = data.error || 'Login gagal. Cek kredensial Anda.';
        }
    } catch (error) {
        msgElement.textContent = 'Terjadi kesalahan saat terhubung ke server.';
        console.error('Login error:', error);
    }
}

function logout() {
    localStorage.removeItem(API_KEY_STORAGE);
    showLoginContainer();
    document.getElementById('file-list-table').getElementsByTagName('tbody')[0].innerHTML = '';
}


// --- FUNGSI CRUD ---

// CREATE
async function uploadFile() {
    const fileInput = document.getElementById('file-input');
    const file = fileInput.files[0];

    if (!file) {
        showToast('Pilih file terlebih dahulu!', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('file', file);

    try {
        const response = await fetch(`${API_URL}/upload`, {
            method: 'POST',
            headers: { 'X-API-Key': getAPIKey() },
            body: formData
        });

        const result = await response.json();

        if (response.ok) {
            showToast(`File berhasil diupload!`, 'success');
            listFiles(); // Refresh daftar setelah upload
            closeUploadModal(); // Close modal after successful upload
            fileInput.value = ''; // Reset input file
        } else {
            showToast(`Gagal upload: ${result.message || 'Unknown error'}`, 'error');
        }
    } catch (error) {
        showToast('Terjadi kesalahan jaringan saat upload.', 'error');
        console.error('Upload error:', error);
    }
}

// Helper function to format file size
function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// READ (List)
async function listFiles() {
    const tbody = document.getElementById('file-list-body');
    tbody.innerHTML = '<tr><td colspan="5" class="px-4 py-3 text-center">Memuat...</td></tr>';

    try {
        const response = await fetch(`${API_URL}/list`, {
            headers: { 'X-API-Key': getAPIKey() }
        });

        if (!response.ok) {
            throw new Error('Gagal memuat daftar file. API Key mungkin tidak valid.');
        }

        allFiles = await response.json();
        currentPage = 1;
        displayFiles();
    } catch (error) {
        tbody.innerHTML = '<tr><td colspan="5" class="px-4 py-3 text-center text-red-500">Error: ' + error.message + '</td></tr>';
        console.error('List error:', error);
    }
}

// Display files based on current page and items per page
function displayFiles() {
    const tbody = document.getElementById('file-list-body');
    tbody.innerHTML = '';

    // Calculate pagination
    const startIndex = (currentPage - 1) * itemsPerPage;
    const endIndex = Math.min(startIndex + itemsPerPage, allFiles.length);
    const filesToDisplay = allFiles.slice(startIndex, endIndex);

    if (filesToDisplay.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="px-4 py-3 text-center">Tidak ada file yang ditemukan.</td></tr>';
        updatePaginationInfo(0, 0, 0);
        updatePaginationControls(1, 1);
        return;
    }

    filesToDisplay.forEach(function(file) {
        const row = tbody.insertRow();
        
        // Original filename cell
        const originalNameCell = row.insertCell();
        originalNameCell.className = 'px-4 py-3 text-sm text-gray-900';
        originalNameCell.textContent = file.original_name;
        
        // Stored filename cell
        const storedNameCell = row.insertCell();
        storedNameCell.className = 'px-4 py-3 text-sm text-gray-900';
        storedNameCell.textContent = file.stored_name;
        
        // Size cell
        const sizeCell = row.insertCell();
        sizeCell.className = 'px-4 py-3 text-sm text-gray-500';
        sizeCell.textContent = formatFileSize(file.file_size);
        
        // Upload time cell
        const timeCell = row.insertCell();
        timeCell.className = 'px-4 py-3 text-sm text-gray-500';
        // Format the upload time to be more readable
        const uploadTime = new Date(file.upload_time);
        timeCell.textContent = uploadTime.toLocaleString('id-ID');
        
        // Actions cell
        const actionsCell = row.insertCell();
        actionsCell.className = 'px-4 py-3 text-sm';
        
        // Preview button
        const previewBtn = document.createElement('button');
        previewBtn.textContent = '👁️';
        previewBtn.className = 'text-green-600 hover:text-green-900 mr-3 font-medium';
        previewBtn.title = 'Preview';
        previewBtn.onclick = function() { previewFile(file.stored_name); };
        actionsCell.appendChild(previewBtn);
        
        // Copy button
        const copyBtn = document.createElement('button');
        copyBtn.textContent = '📋';
        copyBtn.className = 'text-yellow-600 hover:text-yellow-900 mr-3 font-medium';
        copyBtn.title = 'Copy URL';
        copyBtn.onclick = function() { copyFileUrl(file.stored_name); };
        actionsCell.appendChild(copyBtn);
        
        // Download button
        const downloadBtn = document.createElement('button');
        downloadBtn.textContent = '⬇️';
        downloadBtn.className = 'text-blue-600 hover:text-blue-900 mr-3 font-medium';
        downloadBtn.title = 'Download';
        downloadBtn.onclick = function() { downloadFile(file.stored_name); };
        actionsCell.appendChild(downloadBtn);
        
        // Delete button
        const deleteBtn = document.createElement('button');
        deleteBtn.textContent = '🗑️';
        deleteBtn.className = 'text-red-600 hover:text-red-900 font-medium';
        deleteBtn.title = 'Delete';
        deleteBtn.onclick = function() { deleteFile(file.stored_name); };
        actionsCell.appendChild(deleteBtn);
    });

    updatePaginationInfo(startIndex + 1, endIndex, allFiles.length);
    updatePaginationControls(currentPage, Math.ceil(allFiles.length / itemsPerPage));
}

// Update pagination info text
function updatePaginationInfo(start, end, total) {
    document.getElementById('start-item').textContent = start;
    document.getElementById('end-item').textContent = end;
    document.getElementById('total-items').textContent = total;
}

// Update pagination controls
function updatePaginationControls(currentPage, totalPages) {
    const prevBtn = document.getElementById('prev-page');
    const nextBtn = document.getElementById('next-page');
    
    prevBtn.disabled = currentPage <= 1;
    nextBtn.disabled = currentPage >= totalPages;
    
    // Create pagination numbers
    const paginationNumbers = document.getElementById('pagination-numbers');
    paginationNumbers.innerHTML = '';
    
    // Determine which page numbers to show
    let startPage, endPage;
    if (totalPages <= 5) {
        startPage = 1;
        endPage = totalPages;
    } else {
        if (currentPage <= 3) {
            startPage = 1;
            endPage = 5;
        } else if (currentPage >= totalPages - 2) {
            startPage = totalPages - 4;
            endPage = totalPages;
        } else {
            startPage = currentPage - 2;
            endPage = currentPage + 2;
        }
    }
    
    for (let i = startPage; i <= endPage; i++) {
        const pageBtn = document.createElement('button');
        pageBtn.textContent = i;
        pageBtn.className = `px-3 py-1 rounded ${
            i === currentPage 
                ? 'bg-blue-600 text-white' 
                : 'bg-gray-200 hover:bg-gray-300'
        }`;
        pageBtn.onclick = () => goToPage(i);
        paginationNumbers.appendChild(pageBtn);
    }
}

// Change items per page
function changeItemsPerPage() {
    itemsPerPage = parseInt(document.getElementById('items-per-page').value);
    currentPage = 1;
    displayFiles();
}

// Go to specific page
function goToPage(page) {
    if (page < 1 || page > Math.ceil(allFiles.length / itemsPerPage)) {
        return;
    }
    currentPage = page;
    displayFiles();
}

// Go to next page
function goToNextPage() {
    if (currentPage < Math.ceil(allFiles.length / itemsPerPage)) {
        currentPage++;
        displayFiles();
    }
}

// Go to previous page
function goToPrevPage() {
    if (currentPage > 1) {
        currentPage--;
        displayFiles();
    }
}

// READ (Download)
function downloadFile(filename) {
    // Tidak perlu POST/DELETE, cukup GET dan sertakan API Key di header untuk otentikasi
    const url = `${API_URL}/download?name=${encodeURIComponent(filename)}`;
    
    // Karena browser tidak bisa menambahkan header pada navigasi langsung, 
    // kita perlu otentikasi via query parameter atau proxy jika ingin menggunakan API Key.
    // Dalam implementasi ini, kita harus membuat API Key tidak wajib untuk download, 
    // atau menggunakan cara yang lebih kompleks. 
    
    // **ASUMSI SEMENTARA:** Untuk menyederhanakan, kita navigasi langsung. 
    // Ini **TIDAK AMAN** karena API Key tidak terkirim. 
    // Solusi terbaik adalah membuat tautan download memerlukan API Key atau menggunakan JWT.

    // Untuk POC (Proof of Concept) ini, saya akan menggunakan URL yang mengarahkan ke API download
    // Asumsi: Server Go akan mengurus header download, tetapi otentikasi harus dilakukan oleh handler.
    // Karena download via `window.location.href` tidak bisa menyertakan `X-API-Key` header,
    // kita harus menggunakan AJAX/Fetch untuk mendapatkan blob, lalu membuat link download.

    fetch(url, {
        headers: {
            'X-API-Key': getAPIKey() 
        }
    })
    .then(response => {
        if (!response.ok) {
            alert(`Gagal download: ${response.statusText}`);
            return;
        }
        return response.blob();
    })
    .then(blob => {
        const link = document.createElement('a');
        link.href = URL.createObjectURL(blob);
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
    })
    .catch(error => {
        alert('Terjadi kesalahan saat download.');
        console.error('Download error:', error);
    });
}


// Helper function to show toast notifications
function showToast(message, type) {
    // Remove existing toasts
    const existingToasts = document.querySelectorAll('.toast-notification');
    existingToasts.forEach(toast => toast.remove());

    // Create toast container
    const toastContainer = document.createElement('div');
    toastContainer.className = 'fixed top-4 right-4 z-50';
    
    // Create toast element
    const toast = document.createElement('div');
    toast.className = `toast-notification p-4 rounded-md shadow-lg text-white ${
        type === 'success' ? 'bg-green-500' : 'bg-red-500'
    }`;
    toast.textContent = message;
    
    toastContainer.appendChild(toast);
    document.body.appendChild(toastContainer);
    
    // Auto remove after 3 seconds
    setTimeout(() => {
        toastContainer.remove();
    }, 3000);
}

// Preview file in new tab
function previewFile(filename) {
    // Construct the public file URL for preview
    const previewUrl = `${API_URL.replace('/api', '')}/file/${encodeURIComponent(filename)}`;
    
    // Open in new tab
    window.open(previewUrl, '_blank');
}

// Copy file URL to clipboard
async function copyFileUrl(filename) {
    const fileUrl = `${API_URL.replace('/api', '')}/file/${encodeURIComponent(filename)}`;
    
    try {
        await navigator.clipboard.writeText(fileUrl);
        showToast('URL berhasil disalin!', 'success');
    } catch (err) {
        console.error('Failed to copy URL: ', err);
        showToast('Gagal menyalin URL', 'error');
    }
}

// DELETE
async function deleteFile(filename) {
    if (!confirm(`Anda yakin ingin menghapus file: ${filename}?`)) {
        return;
    }

    try {
        const response = await fetch(`${API_URL}/delete`, {
            method: 'DELETE',
            headers: { 
                'Content-Type': 'application/json',
                'X-API-Key': getAPIKey() 
            },
            body: JSON.stringify({ filename })
        });

        const text = await response.text();

        if (response.ok) {
            showToast(`File berhasil dihapus!`, 'success');
            listFiles(); // Refresh daftar setelah delete
        } else {
            showToast(`Gagal menghapus: ${text}`, 'error');
        }
    } catch (error) {
        showToast('Terjadi kesalahan jaringan saat menghapus file.', 'error');
        console.error('Delete error:', error);
    }
}
