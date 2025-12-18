/**
 * Objects Module - Handles all object-related functionality
 */

const ObjectsModule = (function () {
  // Private state
  let nextToken = null;
  let filter = "";
  let pageSize = 20;
  let selectedKeysToDelete = [];

  /**
   * Gets the current bucket name from URL
   * @returns {string} Bucket name
   */
  function getBucketName() {
    return S3Utils.getQueryParam("bucket");
  }

  /**
   * Gets the current path from URL
   * @returns {string} Current path
   */
  function getCurrentPath() {
    return S3Utils.getQueryParam("path") || "";
  }

  /**
   * Initializes the objects page
   */
  function init() {
    const bucket = getBucketName();
    if (!bucket) {
      S3Utils.showToast("No bucket specified", "error");
      return;
    }

    const urlParams = new URLSearchParams(window.location.search);
    filter = urlParams.get("filter") || "";

    // Load page size from URL, then localStorage, then default
    const urlPageSize = parseInt(urlParams.get("count"));
    const localPageSize = parseInt(localStorage.getItem("s3manager_page_size"));
    pageSize = urlPageSize || localPageSize || 20;

    if (pageSize) {
      localStorage.setItem("s3manager_page_size", pageSize);
    }

    const filterInput = document.getElementById("object-filter");
    if (filterInput) {
      filterInput.value = filter;
    }

    const pageSizeSelect = document.getElementById("object-page-size");
    if (pageSizeSelect) {
      pageSizeSelect.value = pageSize;
    }

    // Set bucket title
    const bucketTitle = document.getElementById("bucket-title");
    if (bucketTitle) {
      bucketTitle.textContent = bucket;
    }

    setupNavigation();
    setupEventListeners();
    loadObjects(true);
  }

  /**
   * Sets up back navigation
   */
  function setupNavigation() {
    const path = getCurrentPath();
    const bucket = getBucketName();
    const backButton = document.getElementById("back-button");
    const currentPathEl = document.getElementById("current-path");

    // Create breadcrumb navigation
    if (currentPathEl) {
      let html = `<a href="objects.html?bucket=${encodeURIComponent(bucket)}&path=">/</a>`;
      if (path !== "") {
        let current = "";
        const parts = path.split("/");
        parts.forEach((part, index) => {
          current += (current ? "/" : "") + part;
          if (index < parts.length - 1) {
            html += ` <a href="objects.html?bucket=${encodeURIComponent(bucket)}&path=${encodeURIComponent(current)}">${S3Utils.escapeHtml(part)}</a> /`;
          } else {
            html += ` ${S3Utils.escapeHtml(part)}`;
          }
        });
      }
      currentPathEl.innerHTML = html;
    }

    if (path === "") {
      // At root level - back to buckets list
      if (backButton) {
        backButton.href = "index.html";
        backButton.style.display = "inline-flex";
        const btnText = backButton.querySelector(".btn-text");
        if (btnText) btnText.textContent = "Back";
      }
    } else {
      // In a subfolder
      const prevPath = path.includes("/")
        ? path.slice(0, path.lastIndexOf("/"))
        : "";
      if (backButton) {
        backButton.href = `objects.html?bucket=${encodeURIComponent(bucket)}&path=${encodeURIComponent(prevPath)}`;
        backButton.style.display = "inline-flex";
        const btnText = backButton.querySelector(".btn-text");
        if (btnText) btnText.textContent = "Up";
      }
    }
  }

  /**
   * Sets up all event listeners for the objects page
   */
  function setupEventListeners() {
    // Upload form
    const uploadForm = document.getElementById("upload-form");
    if (uploadForm) {
      uploadForm.addEventListener("submit", handleUpload);
    }

    // Filter form
    const filterForm = document.getElementById("object-filter-form");
    if (filterForm) {
      filterForm.addEventListener("submit", handleFilter);
    }

    // Page size selector
    const pageSizeSelect = document.getElementById("object-page-size");
    if (pageSizeSelect) {
      pageSizeSelect.addEventListener("change", handlePageSizeChange);
    }

    // Load more button
    const loadMoreBtn = document.getElementById("load-more-objects");
    if (loadMoreBtn) {
      loadMoreBtn.addEventListener("click", () => loadObjects(false));
    }

    // Select all checkbox
    const selectAll = document.getElementById("select-all");
    if (selectAll) {
      selectAll.addEventListener("change", handleSelectAll);
    }

    // Delete selected button
    const deleteSelectedBtn = document.getElementById("delete-selected");
    if (deleteSelectedBtn) {
      deleteSelectedBtn.addEventListener("click", showDeleteSelectedModal);
    }

    // Download selected button
    const downloadSelectedBtn = document.getElementById("download-selected");
    if (downloadSelectedBtn) {
      downloadSelectedBtn.addEventListener("click", handleDownloadSelected);
    }

    // Modal buttons
    const cancelDeleteBtn = document.getElementById("cancel-delete-object");
    if (cancelDeleteBtn) {
      cancelDeleteBtn.addEventListener("click", closeDeleteModal);
    }

    const confirmDeleteBtn = document.getElementById("confirm-delete-object");
    if (confirmDeleteBtn) {
      confirmDeleteBtn.addEventListener("click", confirmDelete);
    }

    const cancelDeleteSelectedBtn = document.getElementById(
      "cancel-delete-selected",
    );
    if (cancelDeleteSelectedBtn) {
      cancelDeleteSelectedBtn.addEventListener(
        "click",
        closeDeleteSelectedModal,
      );
    }

    const confirmDeleteSelectedBtn = document.getElementById(
      "confirm-delete-selected",
    );
    if (confirmDeleteSelectedBtn) {
      confirmDeleteSelectedBtn.addEventListener("click", confirmDeleteSelected);
    }
  }

  /**
   * Loads objects from the API
   * @param {boolean} reset - Whether to reset the list
   */
  async function loadObjects(reset = true) {
    const tbody = document.querySelector("#objects-table tbody");
    if (!tbody) return;

    if (reset) {
      nextToken = null;
      tbody.innerHTML = "";
    }

    const bucket = getBucketName();
    const path = getCurrentPath();

    // Show loading state
    const table = document.getElementById("objects-table");
    S3Utils.showLoading(table);

    // Update URL with filter and count
    const url = new URL(window.location);
    if (filter) {
      url.searchParams.set("filter", filter);
    } else {
      url.searchParams.delete("filter");
    }
    url.searchParams.set("count", pageSize);
    window.history.replaceState({}, "", url);

    try {
      const data = await S3API.get(`/buckets/${bucket}`, {
        path: path,
        count: pageSize,
        filter: filter,
        token: nextToken,
      });

      const objects = data.list || [];
      renderObjects(objects, tbody, bucket, path);
      nextToken = data.next_token || null;

      // Toggle load more button
      const loadMoreBtn = document.getElementById("load-more-objects");
      if (loadMoreBtn) {
        loadMoreBtn.style.display = nextToken ? "block" : "none";
      }

      // Reset select all
      const selectAll = document.getElementById("select-all");
      if (selectAll && reset) {
        selectAll.checked = false;
      }
    } catch (error) {
      S3Utils.showToast(`Error loading objects: ${error.message}`);
    } finally {
      S3Utils.hideLoading(table);
    }
  }

  /**
   * Renders objects to the table
   * @param {Array} objects - Array of object data
   * @param {HTMLElement} tbody - Table body element
   * @param {string} bucket - Bucket name
   * @param {string} path - Current path
   */
  function renderObjects(objects, tbody, bucket, path) {
    if (!objects || objects.length === 0) {
      const tr = document.createElement("tr");
      tr.innerHTML = `
                <td colspan="5" class="empty-state">
                    <div class="empty-state-content">
                        <span class="empty-state-icon">📄</span>
                        <p>No objects found</p>
                        <p class="text-muted">Upload files to get started</p>
                    </div>
                </td>`;
      tbody.appendChild(tr);
      return;
    }

    objects.forEach((obj) => {
      const fullKey = path === "" ? obj.key : `${path}/${obj.key}`;
      const tr = document.createElement("tr");

      const checkboxCell = obj.is_dir
        ? `<td><input type="checkbox" disabled title="Cannot select folders"></td>`
        : `<td><input type="checkbox" class="select-object" value="${S3Utils.escapeHtml(obj.key)}"></td>`;

      const nameCell = obj.is_dir
        ? `<td>
                    <div class="item-name">
                        <span class="item-icon">📁</span>
                        <a href="objects.html?bucket=${encodeURIComponent(bucket)}&path=${encodeURIComponent(fullKey)}" class="item-link">
                            ${S3Utils.escapeHtml(obj.key)}/
                        </a>
                    </div>
                </td>`
        : `<td>
                    <div class="item-name">
                        <span class="item-icon">${getFileIcon(obj.key)}</span>
                        <span>${S3Utils.escapeHtml(obj.key)}</span>
                    </div>
                </td>`;

      const actionButton = obj.is_dir
        ? `<button class="btn btn-primary btn-sm" onclick="ObjectsModule.openFolder('${S3Utils.escapeHtml(bucket)}', '${S3Utils.escapeHtml(fullKey)}')">
                        <span class="btn-icon">📂</span>
                        <span class="btn-text">Open</span>
                    </button>`
        : `<button class="btn btn-primary btn-sm" onclick="ObjectsModule.downloadObject('${S3Utils.escapeHtml(bucket)}', '${S3Utils.escapeHtml(fullKey)}')">
                        <span class="btn-icon">⬇</span>
                        <span class="btn-text">Download</span>
                    </button>`;

      tr.innerHTML = `
                ${checkboxCell}
                ${nameCell}
                <td class="cell-size">${S3Utils.formatFileSize(obj.size)}</td>
                <td class="cell-date">${S3Utils.formatDate(obj.last_modified)}</td>
                <td class="cell-actions">
                    <div class="action-buttons">
                        ${actionButton}
                        <button class="btn btn-danger btn-sm" onclick="ObjectsModule.deleteObject('${S3Utils.escapeHtml(bucket)}', '${S3Utils.escapeHtml(fullKey)}', ${obj.is_dir})">
                            <span class="btn-icon">🗑</span>
                            <span class="btn-text">Delete</span>
                        </button>
                    </div>
                </td>`;
      tbody.appendChild(tr);
    });
  }

  /**
   * Gets an appropriate icon for a file type
   * @param {string} filename - File name
   * @returns {string} Emoji icon
   */
  function getFileIcon(filename) {
    const ext = filename.split(".").pop().toLowerCase();
    const icons = {
      // Images
      jpg: "🖼",
      jpeg: "🖼",
      png: "🖼",
      gif: "🖼",
      svg: "🖼",
      webp: "🖼",
      // Documents
      pdf: "📕",
      doc: "📘",
      docx: "📘",
      txt: "📄",
      md: "📝",
      // Code
      js: "📜",
      ts: "📜",
      py: "🐍",
      go: "🔷",
      rs: "🦀",
      html: "🌐",
      css: "🎨",
      // Data
      json: "📋",
      xml: "📋",
      csv: "📊",
      xlsx: "📊",
      xls: "📊",
      // Archives
      zip: "📦",
      tar: "📦",
      gz: "📦",
      rar: "📦",
      "7z": "📦",
      // Media
      mp3: "🎵",
      wav: "🎵",
      mp4: "🎬",
      avi: "🎬",
      mkv: "🎬",
      // Config
      yaml: "⚙",
      yml: "⚙",
      toml: "⚙",
      ini: "⚙",
      env: "⚙",
    };
    return icons[ext] || "📄";
  }

  /**
   * Handles upload form submission
   * @param {Event} e - Submit event
   */
  async function handleUpload(e) {
    e.preventDefault();

    const fileInput = document.getElementById("file-input");
    const folderInput = document.getElementById("folder-input");
    const files = [...(fileInput?.files || []), ...(folderInput?.files || [])];

    if (files.length === 0) {
      S3Utils.showToast("No files selected", "warning");
      return;
    }

    const submitBtn = e.target.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.setAttribute("aria-busy", "true");

    const bucket = getBucketName();
    const path = getCurrentPath();
    let successCount = 0;
    let errorCount = 0;

    for (const file of files) {
      const formData = new FormData();
      let key = file.webkitRelativePath || file.name;
      if (path) {
        key = `${path}/${key}`;
      }
      formData.append("key", key);
      formData.append("file", file);

      try {
        await S3API.putFormData(`/buckets/${bucket}/objects`, formData);
        successCount++;
      } catch (error) {
        errorCount++;
        console.error(`Failed to upload ${file.name}:`, error);
      }
    }

    // Clear inputs
    if (fileInput) fileInput.value = "";
    if (folderInput) folderInput.value = "";

    submitBtn.disabled = false;
    submitBtn.setAttribute("aria-busy", "false");

    if (errorCount > 0) {
      S3Utils.showToast(
        `Uploaded ${successCount} files, ${errorCount} failed`,
        "warning",
      );
    } else {
      S3Utils.showToast(
        `Successfully uploaded ${successCount} file(s)`,
        "success",
      );
    }

    loadObjects(true);
  }

  /**
   * Handles filter form submission
   * @param {Event} e - Submit event
   */
  function handleFilter(e) {
    e.preventDefault();
    filter = document.getElementById("object-filter").value.trim();
    loadObjects(true);
  }

  /**
   * Handles page size change
   * @param {Event} e - Change event
   */
  function handlePageSizeChange(e) {
    pageSize = Math.min(parseInt(e.target.value, 10), 1000);
    localStorage.setItem("s3manager_page_size", pageSize);
    loadObjects(true);
  }

  /**
   * Handles select all checkbox change
   * @param {Event} e - Change event
   */
  function handleSelectAll(e) {
    document.querySelectorAll(".select-object").forEach((cb) => {
      cb.checked = e.target.checked;
    });
  }

  /**
   * Opens a folder
   * @param {string} bucket - Bucket name
   * @param {string} path - Folder path
   */
  function openFolder(bucket, path) {
    window.location.href = `objects.html?bucket=${encodeURIComponent(bucket)}&path=${encodeURIComponent(path)}`;
  }

  /**
   * Downloads an object
   * @param {string} bucket - Bucket name
   * @param {string} key - Object key
   */
  function downloadObject(bucket, key) {
    window.open(S3API.getObjectDownloadUrl(bucket, key), "_blank");
  }

  /**
   * Shows delete confirmation modal for a single object
   * @param {string} bucket - Bucket name
   * @param {string} key - Object key
   * @param {boolean} isDir - Whether the object is a directory
   */
  function deleteObject(bucket, key, isDir) {
    const modal = document.getElementById("delete-object-modal");
    const nameSpan = document.getElementById("delete-object-name");
    const recursiveContainer = document.getElementById(
      "delete-object-recursive-container",
    );
    const recursiveCheckbox = document.getElementById(
      "delete-object-recursive",
    );

    if (nameSpan) nameSpan.textContent = key;

    if (recursiveContainer) {
      recursiveContainer.style.display = isDir ? "block" : "none";
    }
    if (recursiveCheckbox) {
      recursiveCheckbox.checked = false;
    }

    if (modal) modal.showModal();
  }

  /**
   * Closes the delete confirmation modal
   */
  function closeDeleteModal() {
    const modal = document.getElementById("delete-object-modal");
    if (modal) modal.close();
  }

  /**
   * Confirms and executes object deletion
   */
  async function confirmDelete() {
    const bucket = getBucketName();
    const key = document.getElementById("delete-object-name").textContent;
    const recursive =
      document.getElementById("delete-object-recursive")?.checked || false;

    closeDeleteModal();

    try {
      const params = recursive ? { recursive: "true" } : {};
      await S3API.delete(
        `/buckets/${bucket}/objects/${encodeURIComponent(key)}`,
        params,
      );
      loadObjects(true);
      S3Utils.showToast(`Object "${key}" was deleted`, "success");
    } catch (error) {
      S3Utils.showToast(`Error deleting object: ${error.message}`);
    }
  }

  /**
   * Shows the delete selected modal
   */
  function showDeleteSelectedModal() {
    const checkboxes = document.querySelectorAll(".select-object:checked");
    const keys = [...checkboxes].map((cb) => cb.value);

    if (keys.length === 0) {
      S3Utils.showToast("No objects selected", "warning");
      return;
    }

    selectedKeysToDelete = keys;
    const countSpan = document.getElementById("delete-selected-count");
    if (countSpan) countSpan.textContent = keys.length;

    const modal = document.getElementById("delete-selected-modal");
    if (modal) modal.showModal();
  }

  /**
   * Closes the delete selected modal
   */
  function closeDeleteSelectedModal() {
    const modal = document.getElementById("delete-selected-modal");
    if (modal) modal.close();
  }

  /**
   * Confirms and executes deletion of selected objects
   */
  async function confirmDeleteSelected() {
    const bucket = getBucketName();
    const path = getCurrentPath();

    closeDeleteSelectedModal();

    let successCount = 0;
    let errorCount = 0;

    for (const key of selectedKeysToDelete) {
      const fullKey = path === "" ? key : `${path}/${key}`;
      try {
        await S3API.delete(
          `/buckets/${bucket}/objects/${encodeURIComponent(fullKey)}`,
        );
        successCount++;
      } catch (error) {
        errorCount++;
        console.error(`Failed to delete ${key}:`, error);
      }
    }

    // Reset select all
    const selectAll = document.getElementById("select-all");
    if (selectAll) selectAll.checked = false;

    loadObjects(true);

    if (errorCount > 0) {
      S3Utils.showToast(
        `Deleted ${successCount} objects, ${errorCount} failed`,
        "warning",
      );
    } else {
      S3Utils.showToast(`${successCount} object(s) deleted`, "success");
    }
  }

  /**
   * Handles download of selected objects
   */
  function handleDownloadSelected() {
    const checkboxes = document.querySelectorAll(".select-object:checked");
    const keys = [...checkboxes].map((cb) => cb.value);

    if (keys.length === 0) {
      S3Utils.showToast("No objects selected", "warning");
      return;
    }

    const bucket = getBucketName();
    const path = getCurrentPath();

    keys.forEach((key) => {
      const fullKey = path === "" ? key : `${path}/${key}`;
      const a = document.createElement("a");
      a.href = S3API.getObjectDownloadUrl(bucket, fullKey);
      a.download = key;
      a.click();
    });

    S3Utils.showToast(`Downloading ${keys.length} file(s)`, "success");
  }

  // Public API
  return {
    init,
    loadObjects,
    openFolder,
    downloadObject,
    deleteObject,
    closeDeleteModal,
    confirmDelete,
    closeDeleteSelectedModal,
    confirmDeleteSelected,
  };
})();

// Make available globally
window.ObjectsModule = ObjectsModule;

// Legacy support for inline onclick handlers
window.openBucket = ObjectsModule.openFolder;
window.downloadObject = ObjectsModule.downloadObject;
window.deleteObject = ObjectsModule.deleteObject;
window.closeDeleteObjectModal = ObjectsModule.closeDeleteModal;
window.confirmDeleteObject = ObjectsModule.confirmDelete;
window.closeDeleteSelectedModal = ObjectsModule.closeDeleteSelectedModal;
window.confirmDeleteSelected = ObjectsModule.confirmDeleteSelected;
window.loadObjects = function () {
  ObjectsModule.init();
};
