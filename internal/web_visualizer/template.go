package web_visualizer

import (
	"html/template"
)

// GetHTMLTemplate はHTMLテンプレートを返します
func GetHTMLTemplate() (*template.Template, error) {
	// HTMLテンプレートを定義
	const htmlTemplate = `<!DOCTYPE html>
<html lang="ja">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.RepoName}} リポジトリヒートマップ</title>
  <script src="https://d3js.org/d3.v7.min.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/lodash@4.17.21/lodash.min.js"></script>
  <style>
    :root {
      --bg-color: #f8f9fa;
      --text-color: #212529;
      --link-color: #0d6efd;
      --hover-color: #0a58ca;
      --tooltip-bg: rgba(0, 0, 0, 0.8);
      --tooltip-text: #fff;
      --card-bg: #fff;
      --border-color: #dee2e6;
    }

    body {
      font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
      line-height: 1.6;
      color: var(--text-color);
      background-color: var(--bg-color);
      margin: 0;
      padding: 0;
    }

    .container {
      max-width: 1200px;
      margin: 0 auto;
      padding: 2rem;
    }

    header {
      margin-bottom: 2rem;
      text-align: center;
    }

    h1 {
      font-size: 2.5rem;
      margin-bottom: 0.5rem;
    }

    .stats {
      font-size: 1rem;
      color: #6c757d;
      margin-bottom: 1rem;
    }

    .controls {
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.5rem;
      gap: 1rem;
      background: var(--card-bg);
      padding: 1rem;
      border-radius: 0.5rem;
      box-shadow: 0 2px 5px rgba(0,0,0,0.1);
    }

    .control-group {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    select, input {
      padding: 0.5rem;
      border: 1px solid var(--border-color);
      border-radius: 0.25rem;
      font-size: 0.9rem;
    }

    button {
      background-color: var(--link-color);
      color: white;
      border: none;
      padding: 0.5rem 1rem;
      border-radius: 0.25rem;
      cursor: pointer;
      font-size: 0.9rem;
      transition: background-color 0.2s;
    }

    button:hover {
      background-color: var(--hover-color);
    }

    .search {
      flex-grow: 1;
      max-width: 300px;
    }

    input[type="search"] {
      width: 100%;
    }

    .visualization {
      background: var(--card-bg);
      border-radius: 0.5rem;
      box-shadow: 0 2px 5px rgba(0,0,0,0.1);
      padding: 1rem;
      overflow: hidden;
    }

    .treemap {
      width: 100%;
      height: 600px;
      position: relative;
      overflow: auto;
    }

    .tooltip {
      position: absolute;
      padding: 0.75rem;
      background: var(--tooltip-bg);
      color: var(--tooltip-text);
      border-radius: 0.25rem;
      pointer-events: none;
      font-size: 0.9rem;
      z-index: 10;
      max-width: 300px;
      box-shadow: 0 2px 5px rgba(0,0,0,0.2);
    }

    .legend {
      display: flex;
      align-items: center;
      justify-content: center;
      margin-top: 1rem;
      flex-wrap: wrap;
      gap: 1rem;
    }

    .legend-item {
      display: flex;
      align-items: center;
      margin-right: 1rem;
    }

    .legend-color {
      width: 20px;
      height: 12px;
      margin-right: 5px;
      border: 1px solid rgba(0,0,0,0.1);
    }

    .details-panel {
      margin-top: 2rem;
      background: var(--card-bg);
      border-radius: 0.5rem;
      box-shadow: 0 2px 5px rgba(0,0,0,0.1);
      padding: 1rem;
      display: none;
    }

    .details-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1rem;
    }

    .details-content {
      max-height: 400px;
      overflow-y: auto;
      font-family: monospace;
      background-color: #f1f1f1;
      padding: 1rem;
      border-radius: 0.25rem;
    }

    .file-line {
      display: flex;
    }

    .line-number {
      min-width: 3rem;
      text-align: right;
      padding-right: 1rem;
      color: #6c757d;
      user-select: none;
    }

    .breadcrumb {
      display: flex;
      flex-wrap: wrap;
      padding: 0.75rem 1rem;
      margin-bottom: 1rem;
      list-style: none;
      background-color: #e9ecef;
      border-radius: 0.25rem;
    }

    .breadcrumb-item+.breadcrumb-item {
      padding-left: 0.5rem;
    }

    .breadcrumb-item+.breadcrumb-item::before {
      display: inline-block;
      padding-right: 0.5rem;
      content: "/";
    }

    @media (max-width: 768px) {
      .container {
        padding: 1rem;
      }
      
      .controls {
        flex-direction: column;
        align-items: stretch;
      }
      
      .search {
        max-width: 100%;
      }
      
      .treemap {
        height: 400px;
      }
    }

    /* ダークモードサポート */
    @media (prefers-color-scheme: dark) {
      :root {
        --bg-color: #121212;
        --text-color: #e0e0e0;
        --link-color: #90caf9;
        --hover-color: #42a5f5;
        --tooltip-bg: rgba(255, 255, 255, 0.9);
        --tooltip-text: #000;
        --card-bg: #1e1e1e;
        --border-color: #333;
      }
    }
  </style>
</head>
<body>
  <div class="container">
    <header>
      <h1>{{.RepoName}} リポジトリヒートマップ</h1>
      <div class="stats">
        コミット数: {{.JSONData.CommitCount}} / ファイル数: {{.JSONData.FileCount}} / 期間: {{.FirstCommit}} 〜 {{.LastCommit}}
      </div>
    </header>

    <div class="controls">
      <div class="control-group">
        <label for="sort-by">並び替え:</label>
        <select id="sort-by">
          <option value="heat">変更頻度</option>
          <option value="name">名前</option>
          <option value="path">パス</option>
          <option value="date">最終更新日</option>
        </select>
      </div>

      <div class="control-group">
        <label for="max-files">表示件数:</label>
        <select id="max-files">
          <option value="50">50件</option>
          <option value="100">100件</option>
          <option value="200">200件</option>
          <option value="500">500件</option>
        </select>
      </div>

      <div class="search">
        <input type="search" id="search-files" placeholder="ファイル名で検索...">
      </div>

      <button id="toggle-view">ツリー表示 / リスト表示</button>
    </div>

    <nav aria-label="breadcrumb">
      <ol class="breadcrumb" id="path-breadcrumb">
        <li class="breadcrumb-item"><a href="#" data-path="">リポジトリルート</a></li>
      </ol>
    </nav>

    <div class="visualization">
      <div id="treemap" class="treemap"></div>
      
      <div class="legend">
        <span>変更頻度:</span>
        <div class="legend-item">
          <div class="legend-color" style="background-color: #0000FF;"></div>
          <span>低</span>
        </div>
        <div class="legend-item">
          <div class="legend-color" style="background-color: #0066FF;"></div>
        </div>
        <div class="legend-item">
          <div class="legend-color" style="background-color: #00FF00;"></div>
        </div>
        <div class="legend-item">
          <div class="legend-color" style="background-color: #FFCC00;"></div>
        </div>
        <div class="legend-item">
          <div class="legend-color" style="background-color: #FF0000;"></div>
          <span>高</span>
        </div>
      </div>
    </div>

    <div id="details-panel" class="details-panel">
      <div class="details-header">
        <h2 id="details-title">ファイル詳細</h2>
        <button id="close-details">閉じる</button>
      </div>
      <div id="details-content" class="details-content"></div>
    </div>
  </div>

  <script>
    // データを取り込む
    const repoData = {
      files: {{.JSONData.Files}},
      repository_name: "{{.RepoName}}",
      commit_count: {{.JSONData.CommitCount}},
      file_count: {{.JSONData.FileCount}},
      first_commit_at: "{{.FirstCommit}}",
      last_commit_at: "{{.LastCommit}}"
    };
    
    // ファイルツリーを構築する関数
    function buildFileTree(files, maxFiles) {
      // ルートノードを作成
      const root = {
        name: "root",
        path: "",
        children: [],
        isDir: true
      };
      
      // ファイルをソート（変更頻度の高い順）
      const sortedFiles = Object.entries(files)
        .map(function(entry) { 
          const path = entry[0]; 
          const info = entry[1]; 
          return { filePath: path, ...info }; 
        })
        .sort(function(a, b) { return b.changeCount - a.changeCount; });
      
      // 表示するファイル数を制限
      const limitedFiles = sortedFiles.slice(0, maxFiles);
      
      // 各ファイルをツリーに追加
      limitedFiles.forEach(function(file) {
        const pathParts = file.filePath.split(/[\/\\]/);
        let currentNode = root;
        
        // ディレクトリ階層を構築
        for (let i = 0; i < pathParts.length - 1; i++) {
          const dirName = pathParts[i];
          if (!dirName) continue;
          
          // 既存のディレクトリノードを検索
          let found = false;
          for (let j = 0; j < currentNode.children.length; j++) {
            const child = currentNode.children[j];
            if (child.name === dirName && child.isDir) {
              currentNode = child;
              found = true;
              break;
            }
          }
          
          // 見つからなければ新しいディレクトリノードを作成
          if (!found) {
            const dirNode = {
              name: dirName,
              path: pathParts.slice(0, i + 1).join("/"),
              children: [],
              isDir: true
            };
            currentNode.children.push(dirNode);
            currentNode = dirNode;
          }
        }
        
        // ファイルノードを追加
        const fileName = pathParts[pathParts.length - 1];
        const fileNode = {
          name: fileName,
          path: file.filePath,
          size: file.changeCount,
          heatLevel: file.heatLevel,
          lastModified: file.lastModified,
          isDir: false,
          children: []
        };
        currentNode.children.push(fileNode);
      });
      
      // ディレクトリサイズを計算
      calculateDirSize(root);
      
      return root;
    }
    
    // ディレクトリのサイズを計算する関数（再帰的）
    function calculateDirSize(node) {
      if (!node.isDir) {
        return node.size;
      }
      
      let total = 0;
      for (let i = 0; i < node.children.length; i++) {
        total += calculateDirSize(node.children[i]);
      }
      
      node.size = total;
      return total;
    }
    
    // 現在のビュー状態
    let currentView = 'treemap';
    let currentPath = '';
    let maxFilesToShow = 50;
    let currentRoot = null;
    
    // 初期化
    function init() {
      // コントロールのイベントリスナーを設定
      document.getElementById('max-files').addEventListener('change', updateVisualization);
      document.getElementById('sort-by').addEventListener('change', updateVisualization);
      document.getElementById('search-files').addEventListener('input', _.debounce(updateVisualization, 300));
      document.getElementById('toggle-view').addEventListener('click', toggleView);
      document.getElementById('close-details').addEventListener('click', closeDetails);
      
      // パンくずリストの初期アイテムをクリック可能に
      document.querySelector('#path-breadcrumb a').addEventListener('click', function(e) {
        e.preventDefault();
        navigateTo('');
      });
      
      // 初期表示
      maxFilesToShow = parseInt(document.getElementById('max-files').value);
      updateVisualization();
    }
    
    // 可視化の更新
    function updateVisualization() {
      const searchTerm = document.getElementById('search-files').value.toLowerCase();
      const sortBy = document.getElementById('sort-by').value;
      maxFilesToShow = parseInt(document.getElementById('max-files').value);
      
      // ファイルのフィルタリングとソート
      let files = Object.entries(repoData.files)
        .map(function(entry) { 
          const path = entry[0]; 
          const info = entry[1]; 
          return { filePath: path, ...info }; 
        })
        .filter(function(file) {
          if (searchTerm) {
            return file.filePath.toLowerCase().includes(searchTerm);
          }
          return true;
        });
      
      // ソート
      switch (sortBy) {
        case 'heat':
          files.sort(function(a, b) { return b.changeCount - a.changeCount; });
          break;
        case 'name':
          files.sort(function(a, b) {
            const aName = a.filePath.split(/[\/\\]/).pop();
            const bName = b.filePath.split(/[\/\\]/).pop();
            return aName.localeCompare(bName);
          });
          break;
        case 'path':
          files.sort(function(a, b) { return a.filePath.localeCompare(b.filePath); });
          break;
        case 'date':
          files.sort(function(a, b) { return new Date(b.lastModified) - new Date(a.lastModified); });
          break;
      }
      
      // 表示数を制限
      files = files.slice(0, maxFilesToShow);
      
      // ツリー構造を再構築
      const filesObj = {};
      files.forEach(function(file) {
        filesObj[file.filePath] = {
          changeCount: file.changeCount, 
          heatLevel: file.heatLevel,
          lastModified: file.lastModified
        };
      });
      
      const root = buildFileTree(filesObj, maxFilesToShow);
      currentRoot = root;
      
      // 現在のパスに基づいてビューを更新
      if (currentPath === '') {
        renderTreemap(root);
      } else {
        // パスに基づいてサブツリーを見つける
        const pathParts = currentPath.split(/[\/\\]/);
        let node = root;
        
        for (let i = 0; i < pathParts.length; i++) {
          const part = pathParts[i];
          if (!part) continue;
          
          let found = false;
          for (let j = 0; j < node.children.length; j++) {
            const child = node.children[j];
            if (child.name === part && child.isDir) {
              node = child;
              found = true;
              break;
            }
          }
          
          if (!found) {
            // パスが見つからない場合はルートに戻る
            currentPath = '';
            renderTreemap(root);
            updateBreadcrumb();
            return;
          }
        }
        
        renderTreemap(node);
      }
      
      updateBreadcrumb();
    }
    
    // ツリーマップの描画
    function renderTreemap(root) {
      const width = document.getElementById('treemap').clientWidth;
      const height = document.getElementById('treemap').clientHeight;
      
      // 古いSVGをクリア
      d3.select("#treemap").html("");
      
      // ツールチップを作成
      const tooltip = d3.select("#treemap")
        .append("div")
        .attr("class", "tooltip")
        .style("opacity", 0);
      
      // ツリーマップレイアウトを作成
      const treemap = d3.treemap()
        .size([width, height])
        .paddingOuter(3)
        .paddingTop(20)
        .paddingInner(1)
        .round(true);
      
      // 階層構造を作成
      const hierarchy = d3.hierarchy(root)
        .sum(function(d) { return d.isDir ? 0 : d.size; })
        .sort(function(a, b) { return b.value - a.value; });
      
      // ツリーマップレイアウトを適用
      const nodes = treemap(hierarchy);
      
      // SVGを作成
      const svg = d3.select("#treemap")
        .append("svg")
        .attr("width", width)
        .attr("height", height)
        .style("font-family", "sans-serif")
        .style("overflow", "visible");
      
      // ノードを作成
      const node = svg.selectAll("g")
        .data(nodes.descendants())
        .enter()
        .append("g")
        .attr("transform", function(d) { return "translate(" + d.x0 + "," + d.y0 + ")"; })
        .attr("class", function(d) { return d.children ? "directory" : "file"; });
      
      // ディレクトリのタイトル
      node.filter(function(d) { return d.depth === 1; })
        .append("rect")
        .attr("width", function(d) { return d.x1 - d.x0; })
        .attr("height", 20)
        .attr("fill", "#777")
        .attr("opacity", 0.8);
      
      node.filter(function(d) { return d.depth === 1; })
        .append("text")
        .attr("x", 5)
        .attr("y", 15)
        .attr("font-size", "12px")
        .attr("fill", "white")
        .text(function(d) { return d.data.name; });
      
      // ディレクトリに対しては、クリックで中に入れるようにする
      node.filter(function(d) { return d.data.isDir && d.depth > 0; })
        .style("cursor", "pointer")
        .on("click", function(event, d) {
          event.stopPropagation();
          navigateTo(d.data.path);
        });
      
      // ファイルノードの矩形
      node.filter(function(d) { return !d.children; })
        .append("rect")
        .attr("width", function(d) { return Math.max(0, d.x1 - d.x0); })
        .attr("height", function(d) { return Math.max(0, d.y1 - d.y0); })
        .attr("fill", function(d) { return getHeatColor(d.data.heatLevel); })
        .attr("opacity", 0.9)
        .style("cursor", "pointer")
        .on("mouseover", function(event, d) {
          tooltip.transition()
            .duration(200)
            .style("opacity", .9);
          
          const date = new Date(d.data.lastModified);
          const formattedDate = date.getFullYear() + '/' + 
                               (date.getMonth() + 1).toString().padStart(2, '0') + '/' + 
                               date.getDate().toString().padStart(2, '0');
          
          tooltip.html(
            '<strong>' + d.data.name + '</strong><br/>' +
            'パス: ' + d.data.path + '<br/>' +
            '変更回数: ' + d.data.size + '<br/>' +
            '最終更新: ' + formattedDate
          )
            .style("left", (event.pageX - document.getElementById('treemap').getBoundingClientRect().left + 10) + "px")
            .style("top", (event.pageY - document.getElementById('treemap').getBoundingClientRect().top - 28) + "px");
        })
        .on("mouseout", function() {
          tooltip.transition()
            .duration(500)
            .style("opacity", 0);
        })
        .on("click", function(event, d) {
          event.stopPropagation();
          showFileDetails(d.data);
        });
      
      // ファイル名のラベル（十分なスペースがある場合のみ）
      node.filter(function(d) { return !d.children && (d.x1 - d.x0 > 40) && (d.y1 - d.y0 > 14); })
        .append("text")
        .attr("x", 3)
        .attr("y", 13)
        .attr("font-size", "11px")
        .text(function(d) {
          const name = d.data.name;
          // 長すぎる名前を短縮
          return name.length > 20 ? name.substring(0, 17) + "..." : name;
        })
        .attr("pointer-events", "none")
        .attr("fill", "#000");
      
      // SVGの外側をクリックしたらルートに戻る
      svg.on("click", function() {
        if (currentPath !== '') {
          // 現在のパスの親ディレクトリへ
          const pathParts = currentPath.split(/[\/\\]/);
          pathParts.pop();
          navigateTo(pathParts.join('/'));
        }
      });
    }
    
    // 特定のパスに移動
    function navigateTo(path) {
      currentPath = path;
      updateVisualization();
    }
    
    // パンくずリストの更新
    function updateBreadcrumb() {
      const breadcrumb = document.getElementById('path-breadcrumb');
      
      // ルートアイテム以外をクリア
      while (breadcrumb.children.length > 1) {
        breadcrumb.removeChild(breadcrumb.lastChild);
      }
      
      // 現在のパスが空でなければ、パスの各部分に対応するアイテムを追加
      if (currentPath) {
        const pathParts = currentPath.split(/[\/\\]/);
        let currentPathBuilder = '';
        
        for (let i = 0; i < pathParts.length; i++) {
          const part = pathParts[i];
          if (!part) continue;
          
          currentPathBuilder += (currentPathBuilder ? '/' : '') + part;
          
          const li = document.createElement('li');
          li.className = 'breadcrumb-item';
          
          const a = document.createElement('a');
          a.href = '#';
          a.textContent = part;
          a.dataset.path = currentPathBuilder;
          a.addEventListener('click', function(e) {
            e.preventDefault();
            navigateTo(this.dataset.path);
          });
          
          li.appendChild(a);
          breadcrumb.appendChild(li);
        }
      }
    }
    
    // ファイルの詳細表示
    function showFileDetails(fileData) {
      const detailsPanel = document.getElementById('details-panel');
      const detailsTitle = document.getElementById('details-title');
      const detailsContent = document.getElementById('details-content');
      
      detailsTitle.textContent = fileData.name;
      detailsContent.innerHTML = '<div class="loading">ファイル内容を表示できません（静的HTMLのため）</div>';
      detailsPanel.style.display = 'block';
    }
    
    // 詳細パネルを閉じる
    function closeDetails() {
      document.getElementById('details-panel').style.display = 'none';
    }
    
    // ビュー切り替え（ツリー/リスト）
    function toggleView() {
      currentView = currentView === 'treemap' ? 'list' : 'treemap';
      updateVisualization();
    }
    
    // ヒートマップの色を取得する関数
    function getHeatColor(heatLevel) {
      // ヒートレベルに応じた色を返す（0.0-1.0）
      if (heatLevel < 0.2) return "#0000FF";      // 低（青）
      else if (heatLevel < 0.4) return "#0066FF"; // やや低（水色）
      else if (heatLevel < 0.6) return "#00FF00"; // 中（緑）
      else if (heatLevel < 0.8) return "#FFCC00"; // やや高（黄色）
      else return "#FF0000";                      // 高（赤）
    }
    
    // ページ読み込み完了時に初期化
    document.addEventListener('DOMContentLoaded', init);
  </script>
</body>
</html>`

	// テンプレートをパース
	tmpl, err := template.New("heatmap").Parse(htmlTemplate)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
