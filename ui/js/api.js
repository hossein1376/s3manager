/**
 * API Module - Handles all HTTP requests to the S3 Manager backend
 */

const API_BASE = `${window.location.origin}/api`;

/**
 * Makes a GET request to the API
 * @param {string} endpoint - API endpoint
 * @param {Object} params - Query parameters
 * @returns {Promise<Object>} Response data
 */
async function apiGet(endpoint, params = {}) {
    const url = new URL(`${API_BASE}${endpoint}`);
    Object.entries(params).forEach(([key, value]) => {
        if (value !== null && value !== undefined && value !== '') {
            url.searchParams.set(key, value);
        }
    });

    const response = await fetch(url);
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP ${response.status}`);
    }
    return response.json();
}

/**
 * Makes a POST request to the API
 * @param {string} endpoint - API endpoint
 * @param {Object} data - Request body data
 * @returns {Promise<Object>} Response data
 */
async function apiPost(endpoint, data) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP ${response.status}`);
    }

    // Some endpoints may return empty response
    const text = await response.text();
    return text ? JSON.parse(text) : {};
}

/**
 * Makes a PUT request with FormData to the API
 * @param {string} endpoint - API endpoint
 * @param {FormData} formData - Form data to upload
 * @returns {Promise<Object>} Response data
 */
async function apiPutFormData(endpoint, formData) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        method: 'PUT',
        body: formData
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP ${response.status}`);
    }

    const text = await response.text();
    return text ? JSON.parse(text) : {};
}

/**
 * Makes a DELETE request to the API
 * @param {string} endpoint - API endpoint
 * @param {Object} params - Query parameters
 * @returns {Promise<void>}
 */
async function apiDelete(endpoint, params = {}) {
    let url = `${API_BASE}${endpoint}`;
    const queryString = new URLSearchParams(params).toString();
    if (queryString) {
        url += `?${queryString}`;
    }

    const response = await fetch(url, { method: 'DELETE' });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || `HTTP ${response.status}`);
    }
}

/**
 * Gets the download URL for an object
 * @param {string} bucket - Bucket name
 * @param {string} key - Object key
 * @returns {string} Download URL
 */
function getObjectDownloadUrl(bucket, key) {
    return `${API_BASE}/buckets/${bucket}/objects/${encodeURIComponent(key)}`;
}

// Export for use in other modules
window.S3API = {
    get: apiGet,
    post: apiPost,
    putFormData: apiPutFormData,
    delete: apiDelete,
    getObjectDownloadUrl
};
