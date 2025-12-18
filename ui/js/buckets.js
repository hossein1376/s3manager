/**
 * Buckets Module - Handles all bucket-related functionality
 */

const BucketsModule = (function () {
  // Private state
  let nextToken = null;
  let filter = "";

  /**
   * Initializes the buckets page
   */
  function init() {
    setupEventListeners();
    loadBuckets(true);
  }

  /**
   * Sets up all event listeners for the buckets page
   */
  function setupEventListeners() {
    // Create bucket form
    const createForm = document.getElementById("create-bucket-form");
    if (createForm) {
      createForm.addEventListener("submit", handleCreateBucket);
    }

    // Filter form
    const filterForm = document.getElementById("bucket-filter-form");
    if (filterForm) {
      filterForm.addEventListener("submit", handleFilter);
    }

    // Load more button
    const loadMoreBtn = document.getElementById("load-more-buckets");
    if (loadMoreBtn) {
      loadMoreBtn.addEventListener("click", () => loadBuckets(false));
    }

    // Delete modal buttons
    const cancelBtn = document.getElementById("cancel-delete-bucket");
    if (cancelBtn) {
      cancelBtn.addEventListener("click", closeDeleteModal);
    }

    const confirmBtn = document.getElementById("confirm-delete-bucket");
    if (confirmBtn) {
      confirmBtn.addEventListener("click", confirmDelete);
    }
  }

  /**
   * Loads buckets from the API
   * @param {boolean} reset - Whether to reset the list
   */
  async function loadBuckets(reset = true) {
    const tbody = document.querySelector("#buckets-table tbody");
    if (!tbody) return;

    if (reset) {
      nextToken = null;
      tbody.innerHTML = "";
    }

    // Show loading state
    const table = document.getElementById("buckets-table");
    S3Utils.showLoading(table);

    try {
      const data = await S3API.get("/buckets", {
        count: 20,
        filter: filter,
        token: nextToken,
      });

      const buckets = data.buckets || [];
      renderBuckets(buckets, tbody);
      nextToken = data.next_token || null;

      // Toggle load more button
      const loadMoreBtn = document.getElementById("load-more-buckets");
      if (loadMoreBtn) {
        loadMoreBtn.style.display = nextToken ? "block" : "none";
      }
    } catch (error) {
      S3Utils.showToast(`Error loading buckets: ${error.message}`);
    } finally {
      S3Utils.hideLoading(table);
    }
  }

  /**
   * Renders buckets to the table
   * @param {Array} buckets - Array of bucket objects
   * @param {HTMLElement} tbody - Table body element
   */
  function renderBuckets(buckets, tbody) {
    if (!buckets || buckets.length === 0) {
      const tr = document.createElement("tr");
      tr.innerHTML = `
                <td colspan="3" class="empty-state">
                    <div class="empty-state-content">
                        <span class="empty-state-icon">📦</span>
                        <p>No buckets found</p>
                        <p class="empty-state-hint">Create a new bucket to get started</p>
                    </div>
                </td>`;
      tbody.appendChild(tr);
      return;
    }

    buckets.forEach((bucket) => {
      const tr = document.createElement("tr");
      tr.innerHTML = `
                <td>
                    <div class="item-name">
                        <span class="item-icon">📁</span>
                        <a href="objects.html?bucket=${encodeURIComponent(bucket.name)}&path=" class="item-link">
                            ${S3Utils.escapeHtml(bucket.name)}
                        </a>
                    </div>
                </td>
                <td class="cell-date">${S3Utils.formatDate(bucket.created_at)}</td>
                <td class="cell-actions">
                    <div class="action-buttons">
                        <button class="btn btn-primary btn-sm" onclick="BucketsModule.openBucket('${S3Utils.escapeHtml(bucket.name)}')">
                            <span class="btn-icon">📂</span>
                            <span class="btn-text">Open</span>
                        </button>
                        <button class="btn btn-danger btn-sm" onclick="BucketsModule.deleteBucket('${S3Utils.escapeHtml(bucket.name)}')">
                            <span class="btn-icon">🗑</span>
                            <span class="btn-text">Delete</span>
                        </button>
                    </div>
                </td>`;
      tbody.appendChild(tr);
    });
  }

  /**
   * Handles create bucket form submission
   * @param {Event} e - Submit event
   */
  async function handleCreateBucket(e) {
    e.preventDefault();

    const nameInput = document.getElementById("bucket-name");
    const name = nameInput.value.trim();

    if (!name) {
      S3Utils.showToast("Please enter a bucket name", "warning");
      return;
    }

    const submitBtn = e.target.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.setAttribute("aria-busy", "true");

    try {
      await S3API.post("/buckets", { name });
      nameInput.value = "";
      loadBuckets(true);
      S3Utils.showToast(`Bucket "${name}" was created`, "success");
    } catch (error) {
      S3Utils.showToast(`Error creating bucket: ${error.message}`);
    } finally {
      submitBtn.disabled = false;
      submitBtn.setAttribute("aria-busy", "false");
    }
  }

  /**
   * Handles filter form submission
   * @param {Event} e - Submit event
   */
  function handleFilter(e) {
    e.preventDefault();
    filter = document.getElementById("bucket-filter").value;
    loadBuckets(true);
  }

  /**
   * Opens a bucket
   * @param {string} name - Bucket name
   */
  function openBucket(name) {
    window.location.href = `objects.html?bucket=${encodeURIComponent(name)}&path=`;
  }

  /**
   * Shows delete confirmation modal
   * @param {string} name - Bucket name to delete
   */
  function deleteBucket(name) {
    const modal = document.getElementById("delete-bucket-modal");
    const nameSpan = document.getElementById("delete-bucket-name");
    const recursiveCheckbox = document.getElementById(
      "delete-bucket-recursive",
    );

    if (nameSpan) nameSpan.textContent = name;
    if (recursiveCheckbox) recursiveCheckbox.checked = false;
    if (modal) modal.showModal();
  }

  /**
   * Closes the delete confirmation modal
   */
  function closeDeleteModal() {
    const modal = document.getElementById("delete-bucket-modal");
    if (modal) modal.close();
  }

  /**
   * Confirms and executes bucket deletion
   */
  async function confirmDelete() {
    const name = document.getElementById("delete-bucket-name").textContent;
    const recursive = document.getElementById(
      "delete-bucket-recursive",
    ).checked;

    closeDeleteModal();

    try {
      const params = recursive ? { recursive: "true" } : {};
      await S3API.delete(`/buckets/${name}`, params);
      loadBuckets(true);
      S3Utils.showToast(`Bucket "${name}" was deleted`, "success");
    } catch (error) {
      S3Utils.showToast(`Error deleting bucket: ${error.message}`);
    }
  }

  // Public API
  return {
    init,
    loadBuckets,
    openBucket,
    deleteBucket,
    closeDeleteModal,
    confirmDelete,
  };
})();

// Make available globally
window.BucketsModule = BucketsModule;

// Legacy support for inline onclick handlers
window.openBucket = BucketsModule.openBucket;
window.deleteBucket = BucketsModule.deleteBucket;
window.closeDeleteBucketModal = BucketsModule.closeDeleteModal;
window.confirmDeleteBucket = BucketsModule.confirmDelete;
window.loadBuckets = function () {
  BucketsModule.init();
};
