/**
 * Utility Functions Module
 */

/**
 * Shows a toast notification
 * @param {string} message - Message to display
 * @param {string} type - Type of toast: 'success', 'error', 'warning', 'info'
 * @param {number} timeout - Duration in milliseconds
 */
function showToast(message, type = "error", timeout = 3000) {
  const toast = document.getElementById("toast");
  if (!toast) return;

  // Clear any existing timeout
  if (toast._timeout) {
    clearTimeout(toast._timeout);
  }

  toast.textContent = message;
  toast.className = `toast ${type}`;
  toast.classList.add("show");

  toast._timeout = setTimeout(() => {
    toast.classList.remove("show");
  }, timeout);
}

/**
 * Formats file size to human-readable format
 * @param {number} bytes - Size in bytes
 * @returns {string} Formatted size string
 */
function formatFileSize(bytes) {
  if (bytes === undefined || bytes === null) return "-";
  if (bytes === 0) return "0 Bytes";

  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

/**
 * Formats a date string to a locale-friendly format
 * @param {string} dateString - ISO date string
 * @returns {string} Formatted date string
 */
function formatDate(dateString) {
  if (!dateString) return "-";

  try {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return dateString;
  }
}

/**
 * Gets URL query parameter value
 * @param {string} name - Parameter name
 * @returns {string|null} Parameter value or null
 */
function getQueryParam(name) {
  return new URLSearchParams(window.location.search).get(name);
}

/**
 * Escapes HTML special characters to prevent XSS
 * @param {string} str - String to escape
 * @returns {string} Escaped string
 */
function escapeHtml(str) {
  if (!str) return "";
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

/**
 * Debounces a function call
 * @param {Function} func - Function to debounce
 * @param {number} wait - Wait time in milliseconds
 * @returns {Function} Debounced function
 */
function debounce(func, wait) {
  let timeout;
  return function executedFunction(...args) {
    const later = () => {
      clearTimeout(timeout);
      func(...args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

/**
 * Shows a loading spinner in an element
 * @param {HTMLElement} element - Element to show spinner in
 */
function showLoading(element) {
  if (!element) return;
  element.classList.add("loading");
  element.setAttribute("aria-busy", "true");
}

/**
 * Hides loading spinner from an element
 * @param {HTMLElement} element - Element to hide spinner from
 */
function hideLoading(element) {
  if (!element) return;
  element.classList.remove("loading");
  element.setAttribute("aria-busy", "false");
}

/**
 * Creates an element with attributes and children
 * @param {string} tag - HTML tag name
 * @param {Object} attrs - Attributes object
 * @param {Array|string} children - Child elements or text
 * @returns {HTMLElement} Created element
 */
function createElement(tag, attrs = {}, children = []) {
  const el = document.createElement(tag);

  Object.entries(attrs).forEach(([key, value]) => {
    if (key === "className") {
      el.className = value;
    } else if (key === "onclick" || key === "onchange") {
      el[key] = value;
    } else if (key === "dataset") {
      Object.entries(value).forEach(([dataKey, dataValue]) => {
        el.dataset[dataKey] = dataValue;
      });
    } else {
      el.setAttribute(key, value);
    }
  });

  if (typeof children === "string") {
    el.textContent = children;
  } else if (Array.isArray(children)) {
    children.forEach((child) => {
      if (typeof child === "string") {
        el.appendChild(document.createTextNode(child));
      } else if (child instanceof HTMLElement) {
        el.appendChild(child);
      }
    });
  }

  return el;
}

// Export for use in other modules
window.S3Utils = {
  showToast,
  formatFileSize,
  formatDate,
  getQueryParam,
  escapeHtml,
  debounce,
  showLoading,
  hideLoading,
  createElement,
};
