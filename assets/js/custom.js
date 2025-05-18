// Custom JavaScript for Swagger UI

// Add a custom title to the page
document.addEventListener("DOMContentLoaded", function () {
  // Wait for SwaggerUI to be fully loaded
  setTimeout(function () {
    // Add version info
    const info = document.querySelector(".info");
    if (info) {
      const versionDiv = document.createElement("div");
      versionDiv.className = "version-info";
      versionDiv.innerHTML = "<p>API Version: 1.0.0</p>";
      versionDiv.style.marginBottom = "20px";
      versionDiv.style.fontSize = "14px";
      versionDiv.style.color = "#3b4151";
      info.appendChild(versionDiv);
    }
  }, 1000);
});
