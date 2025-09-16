const API_BASE = "http://127.0.0.1:8080/api";

let bucketNextToken = null;
let bucketFilter = "";
let objectNextToken = null;
let objectFilter = "";

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

  bucketNextToken = data.next_token || null;
  document.querySelector("#load-more-buckets").style.display = bucketNextToken ? "block" : "none";

  document.querySelector("#create-bucket-form").onsubmit = async e => {
    e.preventDefault();
    const name = document.querySelector("#bucket-name").value;
    await fetch(`${API_BASE}/buckets`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({name})
    });
    document.querySelector("#bucket-name").value = "";
    loadBuckets(true);
  };

  document.querySelector("#bucket-filter-form").onsubmit = e => {
    e.preventDefault();
    bucketFilter = document.querySelector("#bucket-filter").value;
    loadBuckets(true);
  };

  document.querySelector("#load-more-buckets").onclick = () => loadBuckets(false);
}

async function deleteBucket(name) {
  await fetch(`${API_BASE}/buckets/${name}`, {method: "DELETE"});
  loadBuckets(true);
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
  url.searchParams.set("count", "20");
  if (objectFilter) url.searchParams.set("filter", objectFilter);
  if (objectNextToken) url.searchParams.set("token", objectNextToken);

  const res = await fetch(url);
  const data = await res.json();

  const tbody = document.querySelector("#objects-table tbody");
  data.list.forEach(o => {
    const tr = document.createElement("tr");
    tr.innerHTML = `
      <td>${o.key}</td>
      <td>${o.size}</td>
      <td>${o.last_modified}</td>
      <td>
        <button onclick="downloadObject('${bucket}','${o.key}')">Download</button>
        <button class="outline contrast pico-background-red-600" onclick="deleteObject('${bucket}','${o.key}')">Delete</button>
      </td>`;
    tbody.appendChild(tr);
  });

  objectNextToken = data.next_token || null;
  document.querySelector("#load-more-objects").style.display = objectNextToken ? "block" : "none";

  document.querySelector("#upload-form").onsubmit = async e => {
    e.preventDefault();
    const fileInput = document.querySelector("#file-input");
    const key = document.querySelector("#object-key").value;
    const formData = new FormData();
    formData.append("key", key);
    formData.append("file", fileInput.files[0]);
    await fetch(`${API_BASE}/buckets/${bucket}/objects`, {
      method: "PUT",
      body: formData
    });
    fileInput.value = "";
    document.querySelector("#object-key").value = "";
    loadObjects(true);
  };

  document.querySelector("#object-filter-form").onsubmit = e => {
    e.preventDefault();
    objectFilter = document.querySelector("#object-filter").value;
    loadObjects(true);
  };

  document.querySelector("#load-more-objects").onclick = () => loadObjects(false);
}

async function deleteObject(bucket, key) {
  await fetch(`${API_BASE}/buckets/${bucket}/objects/${encodeURIComponent(key)}`, {method: "DELETE"});
  loadObjects(true);
}

async function downloadObject(bucket, key) {
  window.open(`${API_BASE}/buckets/${bucket}/objects/${encodeURIComponent(key)}`, "_blank");
}
