const container = document.getElementById("container");
const tooltip = document.getElementById("tooltip");

const svg = d3.select("#container")
    .append("svg");

const colorScale = d3.scaleOrdinal(d3.schemeTableau10);

let currentNode;
let navigationStack = [];
let currentView = 'map';

function switchView(view) {
    currentView = view;
    
    // Update buttons
    document.querySelectorAll('.view-toggle button').forEach(btn => btn.classList.remove('active'));
    document.getElementById(`btn-${view}`).classList.add('active');
    
    // Update content
    document.querySelectorAll('.view-content').forEach(el => el.classList.remove('active'));
    document.getElementById(`view-${view}`).classList.add('active');
    
    // Re-render if needed (especially for map resizing)
    if (currentNode) {
        render(currentNode);
    }
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function getDirectChildren(node) {
    if (!node.children) return [];
    
    return node.children.map(child => {
        if (child.type === "directory") {
            const totalSize = child.value || (child.children ? 
                child.children.reduce((sum, c) => sum + (c.value || c.size || 0), 0) : 0);
            return {
                name: child.name,
                path: child.path,
                value: totalSize,
                type: "directory",
                _original: child
            };
        } else {
            return {
                name: child.name,
                path: child.path,
                value: child.value || 0,
                type: "file",
                _original: child
            };
        }
    });
}

function updateBreadcrumbs() {
    const breadcrumbsDiv = document.getElementById("breadcrumbs");
    breadcrumbsDiv.innerHTML = '';
    
    navigationStack.forEach((node, index) => {
        if (index > 0) {
            const separator = document.createElement("span");
            separator.className = "breadcrumb-separator";
            separator.textContent = "›";
            breadcrumbsDiv.appendChild(separator);
        }
        
        const crumb = document.createElement("span");
        crumb.className = index === navigationStack.length - 1 ? 
            "breadcrumb-item current" : "breadcrumb-item";
        crumb.textContent = node.name.split('/').pop() || node.name;
        
        if (index < navigationStack.length - 1) {
            crumb.addEventListener("click", () => {
                navigationStack = navigationStack.slice(0, index + 1);
                render(navigationStack[navigationStack.length - 1]);
                updateInfo(navigationStack[navigationStack.length - 1]);
            });
        }
        
        breadcrumbsDiv.appendChild(crumb);
    });
}

function render(node) {
    currentNode = node;
    
    // Always render list (it's cheap)
    renderList(node);

    // Only render map if visible
    if (currentView === 'map') {
        renderMap(node);
    }
    
    updateBreadcrumbs();
    updateInfo(node);
}

function renderList(node) {
    const tbody = document.querySelector("#file-table tbody");
    tbody.innerHTML = '';
    
    const children = getDirectChildren(node);
    // Sort by size desc
    children.sort((a, b) => b.value - a.value);
    
    const maxSize = children.length > 0 ? children[0].value : 1;

    children.forEach(child => {
        const tr = document.createElement("tr");
        
        const icon = child.type === 'directory' ? '📁' : '📄';
        const percent = (child.value / maxSize) * 100;
        
        tr.innerHTML = `
            <td><span class="type-icon">${icon}</span>${child.name}</td>
            <td>${formatBytes(child.value)}</td>
            <td>${child.type}</td>
            <td>
                <div class="size-bar-bg">
                    <div class="size-bar-fill" style="width: ${percent}%"></div>
                </div>
            </td>
        `;
        
        if (child.type === 'directory') {
            tr.addEventListener("click", () => {
                navigationStack.push(child._original);
                render(child._original);
            });
        }
        
        tbody.appendChild(tr);
    });
}

function renderMap(node) {
    const rect = container.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;
    
    if (width === 0 || height === 0) return; // Hidden
    
    svg.attr("viewBox", `0 0 ${width} ${height}`);
    
    const directChildren = getDirectChildren(node);
    
    const hierarchyData = {
        name: node.name,
        children: directChildren
    };
    
    const root = d3.hierarchy(hierarchyData)
        .sum(d => d.value || 0)
        .sort((a, b) => b.value - a.value);
    
    d3.treemap()
        .size([width, height])
        .padding(2)
        .round(true)
        (root);
    
    const cells = svg.selectAll("g.cell")
        .data(root.leaves(), d => d.data.path);
    
    cells.exit().remove();
    
    const cellsEnter = cells.enter()
        .append("g")
        .attr("class", d => `cell ${d.data.type}`);
    
    cellsEnter.append("rect");
    cellsEnter.append("text");
    
    const cellsMerge = cellsEnter.merge(cells);
    
    cellsMerge
        .attr("transform", d => `translate(${d.x0},${d.y0})`);
    
    cellsMerge.select("rect")
        .attr("width", d => Math.max(0, d.x1 - d.x0))
        .attr("height", d => Math.max(0, d.y1 - d.y0))
        .attr("fill", (d, i) => colorScale(i));
    
    cellsMerge.select("text")
        .attr("x", 4)
        .attr("y", 4)
        .each(function(d) {
            const textEl = d3.select(this);
            textEl.selectAll("*").remove();
            
            const boxWidth = d.x1 - d.x0;
            const boxHeight = d.y1 - d.y0;
            
            if (boxWidth < 30 || boxHeight < 20) return;
            
            const maxChars = Math.floor(boxWidth / 7);
            let displayName = d.data.name;
            if (displayName.length > maxChars) {
                displayName = displayName.substring(0, maxChars - 3) + '...';
            }
            
            textEl.append("tspan")
                .attr("x", 4)
                .attr("dy", "0em")
                .text(displayName);
            
            if (boxHeight > 40) {
                textEl.append("tspan")
                    .attr("x", 4)
                    .attr("dy", "1.2em")
                    .style("font-size", "10px")
                    .style("opacity", "0.8")
                    .text(formatBytes(d.value));
            }
        });
    
    cellsMerge
        .on("click", function(event, d) {
            event.stopPropagation();
            if (d.data.type === "directory") {
                navigationStack.push(d.data._original);
                render(d.data._original);
            }
        })
        .on("mouseover", function(event, d) {
            const fullPath = d.data.path;
            const size = formatBytes(d.value);
            const type = d.data.type === "directory" ? "📁 Directory" : "📄 File";
            
            tooltip.innerHTML = `
                <div><strong>${type}</strong></div>
                <div>${d.data.name}</div>
                <div style="opacity: 0.7; margin-top: 4px;">${size}</div>
            `;
            tooltip.style.opacity = "1";
        })
        .on("mousemove", function(event) {
            const tooltipRect = tooltip.getBoundingClientRect();
            const containerRect = container.getBoundingClientRect();
            
            let left = event.clientX + 15;
            let top = event.clientY - 125;
            
            // If tooltip would go off right edge, position it to the left of cursor
            if (left + tooltipRect.width > window.innerWidth - 10) {
                left = event.clientX - tooltipRect.width - 25;
            }
            
            // If tooltip would go off bottom of container, position it well above cursor
            // Add extra buffer to account for the info panel
            if (top + tooltipRect.height > containerRect.bottom - 100) {
                top = event.clientY - tooltipRect.height - 125;
            }
            
            // Ensure tooltip doesn't go off left edge
            if (left < 10) {
                left = 10;
            }
            
            // Ensure tooltip doesn't go off top edge
            if (top < containerRect.top) {
                top = containerRect.top + 10;
            }
            
            tooltip.style.left = left + "px";
            tooltip.style.top = top + "px";
        })
        .on("mouseout", function() {
            tooltip.style.opacity = "0";
        });
    
    svg.on("click", function(event) {
        if (event.target.tagName === 'svg' && navigationStack.length > 1) {
            navigationStack.pop();
            const parent = navigationStack[navigationStack.length - 1];
            render(parent);
        }
    });
}

function updateInfo(node) {
    const totalSize = node.value || (node.children ? 
        node.children.reduce((sum, c) => sum + (c.value || c.size || 0), 0) : 0);
    const itemCount = node.children?.length || 0;
    
    d3.select("#info").html(`
        <h3>📁 ${node.name}</h3>
        <p><strong>Path:</strong> ${node.path}</p>
        <p><strong>Total Size:</strong> ${formatBytes(totalSize)}</p>
        <p><strong>Items:</strong> ${itemCount}</p>
    `);
}

window.addEventListener("resize", () => {
    if (currentNode) render(currentNode);
});

fetch("/data")
    .then(res => res.json())
    .then(data => {
        currentNode = data;
        navigationStack = [data];
        render(data);
    })
    .catch(error => {
        // Fallback for local file opening
        console.log("Fetch /data failed, trying transformed.json...");
        fetch("transformed.json")
            .then(res => res.json())
            .then(data => {
                currentNode = data;
                navigationStack = [data];
                render(data);
            })
            .catch(err2 => {
                console.error("Error loading data:", error);
                document.getElementById("info").innerHTML = `
                    <h3>❌ Error</h3>
                    <p>Failed to load data. Please make sure the server is running or transformed.json exists.</p>
                    <p style="margin-top: 10px; opacity: 0.7;">Error: ${error.message}</p>
                `;
            });
    });
