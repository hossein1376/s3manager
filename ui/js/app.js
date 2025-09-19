const API_BASE = `${window.location.origin}/api`;

let bucketNextToken = null;
let bucketFilter = "";
let objectNextToken = null;
let objectFilter = "";
let objectPageSize = 20;

async function loadBuckets(reset = true) {
    if (reset) {
        bucketNextToken = null;
        document.querySelector("#buckets-table tbody").innerHTML = "";
    }

    const url = new URL(`${API_BASE}/buckets`);
    url.searchParams.set("count", "20");
    if (bucketFilter) url.searchParams.set("filter", bucketFilter);
    if (bucketNextToken) url.searchParams.set("token", bucketNextToken);

    const res = await fetch(url);
    if (!res.ok) {
        showToast("Error loading buckets: " + await res.text());
        return
    }
    const data = await res.json();

    const tbody = document.querySelector("#buckets-table tbody");
    data.buckets.forEach(b => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
      <td>${b.name}</td>
      <td>${b.created_at}</td>
      <td>
        <button onclick="openBucket('${b.name}')">Open</button>
        <button class="outline contrast pico-background-red-600" onclick="deleteBucket('${b.name}')">Delete</button>
      </td>`;
        tbody.appendChild(tr);
    });
    if (data.buckets.length === 0) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td colspan="4" style="text-align:center; color:#888;">No buckets found</td>`;
        tbody.appendChild(tr);
    }

    bucketNextToken = data.next_token || null;
    document.querySelector("#load-more-buckets").style.display = bucketNextToken ? "block" : "none";

    document.querySelector("#create-bucket-form").onsubmit = async e => {
        e.preventDefault();
        const name = document.querySelector("#bucket-name").value;
        const res = await fetch(`${API_BASE}/buckets`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify({name})
        });
        if (!res.ok) {
            showToast("Error creating bucket: " + await res.text());
            return
        }
        document.querySelector("#bucket-name").value = "";
        loadBuckets(true);
        showToast(`Bucket "${name}" was created`, "success");
    };

    document.querySelector("#bucket-filter-form").onsubmit = e => {
        e.preventDefault();
        bucketFilter = document.querySelector("#bucket-filter").value;
        loadBuckets(true);
    };

    document.querySelector("#load-more-buckets").onclick = () => loadBuckets(false);
}

async function deleteBucket(name) {
    if (!confirm(`Are you sure you want to delete bucket "${name}"?`)) {
        return;
    }
    const res = await fetch(`${API_BASE}/buckets/${name}`, {method: "DELETE"});
    if (!res.ok) {
        showToast("Error deleting bucket: " + await res.text());
        return
    }
    loadBuckets(true);
    showToast(`Bucket "${name}" was deleted`, "success");
}

function openBucket(name) {
    window.location.href = `objects.html?bucket=${encodeURIComponent(name)}`;
}

async function loadObjects(reset = true) {
    const params = new URLSearchParams(window.location.search);
    const bucket = params.get("bucket");
    document.querySelector("#bucket-title").textContent = bucket;

    if (reset) {
        objectNextToken = null;
        document.querySelector("#objects-table tbody").innerHTML = "";
    }

    const url = new URL(`${API_BASE}/buckets/${bucket}`);
    url.searchParams.set("count", objectPageSize);
    if (objectFilter) url.searchParams.set("filter", objectFilter);
    if (objectNextToken) url.searchParams.set("token", objectNextToken);

    const res = await fetch(url);
    if (!res.ok) {
        showToast("Error loading objects: " + await res.text());
        return
    }
    const data = await res.json();

    const tbody = document.querySelector("#objects-table tbody");
    data.list.forEach(o => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
      <td><input type="checkbox" class="select-object" value="${o.key}"></td>
      <td>${o.key}</td>
      <td>${o.size}</td>
      <td>${o.last_modified}</td>
      <td>
        <button onclick="downloadObject('${bucket}','${o.key}')">Download</button>
        <button class="outline contrast pico-background-red-600" onclick="deleteObject('${bucket}','${o.key}')">Delete</button>
      </td>`;
        tbody.appendChild(tr);
    });
    if (data.list.length === 0) {
        const tr = document.createElement("tr");
        tr.innerHTML = `<td colspan="3" style="text-align:center; color:#888;">No objects found</td>`;
        tbody.appendChild(tr);
    }

    objectNextToken = data.next_token || null;
    document.querySelector("#load-more-objects").style.display = objectNextToken ? "block" : "none";

    document.getElementById("object-page-size").onchange = (e) => {
        objectPageSize = Math.min(parseInt(e.target.value, 10), 1000);
        loadObjects(true);
    };

    document.querySelector("#upload-form").onsubmit = async e => {
        e.preventDefault();
        const fileInput = document.querySelector("#file-input");
        const folderInput = document.querySelector("#folder-input");
        const files = [...fileInput.files, ...folderInput.files];
        if (files.length === 0) return showToast("No files selected");

        for (const file of files) {
            const formData = new FormData();
            console.log(file.webkitRelativePath || file.name)
            formData.append("key", file.webkitRelativePath || file.name); // preserves folder path
            formData.append("file", file);
            try {
                await fetch(`${API_BASE}/buckets/${getBucketName()}/objects`, {
                    method: "PUT",
                    body: formData
                });
            } catch (err) {
                showToast(`Failed to upload ${file.name}: ${err.message}`);
            } finally {
                fileInput.value = "";
                folderInput.value = "";
            }
        }
        loadObjects(true);
        showToast(`Successfully uploaded ${files.length} objects!`, "success");
    };

    document.querySelector("#object-filter-form").onsubmit = e => {
        e.preventDefault();
        objectFilter = document.querySelector("#object-filter").value;
        loadObjects(true);
    };

    document.querySelector("#load-more-objects").onclick = () => loadObjects(false);

    document.getElementById("select-all").onchange = (e) => {
        document.querySelectorAll(".select-object").forEach(cb => {
            cb.checked = e.target.checked;
        });
    };

    document.getElementById("delete-selected").onclick = async () => {
        const keys = [...document.querySelectorAll(".select-object:checked")].map(cb => cb.value);
        if (keys.length === 0) return showToast("No objects selected");
        if (!confirm(`Delete ${keys.length} objects?`)) return;

        for (const key of keys) {
            try {
                await fetch(`${API_BASE}/buckets/${getBucketName()}/objects/${encodeURIComponent(key)}`, {method: "DELETE"});
            } catch (err) {
                showToast(`Failed to delete ${key}: ${err.message}`);
            }
        }
        document.getElementById("select-all").checked = false;
        loadObjects(true);
        showToast(`${keys.length} objects deleted`, "success");
    };

    document.getElementById("download-selected").onclick = () => {
        const keys = [...document.querySelectorAll(".select-object:checked")].map(cb => cb.value);
        if (keys.length === 0) return showToast("No objects selected");
        keys.forEach(key => {
            const a = document.createElement("a");
            a.href = `${API_BASE}/buckets/${getBucketName()}/objects/${encodeURIComponent(key)}`;
            a.download = key;
            a.click();
        });
    };
}

async function deleteObject(bucket, key) {
    if (!confirm(`Are you sure you want to delete object "${key}"?`)) {
        return;
    }
    const res = await fetch(`${API_BASE}/buckets/${bucket}/objects/${encodeURIComponent(key)}`, {method: "DELETE"});
    if (!res.ok) {
        showToast("Failed to delete object: " + await res.text());
        return;
    }
    loadObjects(true);
    showToast(`Object "${key}" was deleted`, "success");
}

async function downloadObject(bucket, key) {
    window.open(`${API_BASE}/buckets/${bucket}/objects/${encodeURIComponent(key)}`, "_blank");
}

function getBucketName() {
    return new URLSearchParams(window.location.search).get("bucket");
}

function showToast(message, type = "error", timeout = 3000) {
    const toast = document.getElementById("toast");
    toast.textContent = message;
    toast.className = "toast " + type;
    toast.classList.add("show");
    setTimeout(() => {
        toast.classList.remove("show");
    }, timeout);
}
