"use strict";

// Renders a rectangular maze from the box-drawing text produced by display.Render.
// The first row is the top of the maze; y increases upward, x increases to the right.
// A vertical bridge (╫) uses the bridge.svg layout. A horizontal bridge (╪) is that
// layout rotated 90 degrees clockwise. Bridge passages are not maze walls.
(function (root, factory) {
  const api = factory();
  if (typeof module === "object" && module.exports) {
    module.exports = api;
  }
  if (root) {
    root.mazeSvg = api;
  }
})(typeof globalThis !== "undefined" ? globalThis : this, function () {
  const N = 1;
  const E = 2;
  const S = 4;
  const W = 8;

  const WALLS = {
    "╬": 0,
    "╦": N,
    "╩": S,
    "╠": W,
    "╣": E,
    "╚": W | S,
    "╝": E | S,
    "╔": W | N,
    "╗": E | N,
    "═": N | S,
    "║": E | W,
    "·": N | E | S | W,
    "█": N | E | S | W,
    "╨": E | S | W,
    "╞": N | S | W,
    "╡": N | E | S,
    "╥": N | E | W,
  };

  const BRIDGES = new Set(["╫", "╪"]);

  const CELL = 32;
  const STROKE = 2;
  const PAD = STROKE;
  const BACKGROUND = "#d3d3d3";
  // bridge.svg splits the cell into 11ths: approach, shoulder, deck, shoulder,
  // approach (2, 2, 3, 2, 2). Rails sit one eleventh in from each side.
  const BRIDGE_UNIT = CELL / 11;
  const BRIDGE_FILL = "#feffff";

  const EXIT_PAIR = /\(\s*-?\d+\s*,\s*-?\d+\s*\)\s*-\s*\(\s*-?\d+\s*,\s*-?\d+\s*\)/;

  function renderMazeSvg(lines, entrances) {
    const grid = parseGrid(lines);
    const height = grid.length;
    const width = grid[0].length;
    const exits = normalizeEntrances(entrances, width, height);
    const exitAt = new Set(exits.map((e) => e.x + "," + e.y));

    const hWall = Array.from({ length: height + 1 }, () => Array(width).fill(false));
    const vWall = Array.from({ length: height }, () => Array(width + 1).fill(false));

    for (let row = 0; row < height; row++) {
      for (let x = 0; x < width; x++) {
        const cell = grid[row][x];
        if (cell.bridge) continue;
        if (cell.walls & N) hWall[row][x] = true;
        if (cell.walls & S) hWall[row + 1][x] = true;
        if (cell.walls & W) vWall[row][x] = true;
        if (cell.walls & E) vWall[row][x + 1] = true;
      }
    }

    for (let x = 0; x < width; x++) {
      hWall[0][x] = !exitAt.has(x + "," + (height - 1));
      hWall[height][x] = !exitAt.has(x + ",0");
    }
    for (let row = 0; row < height; row++) {
      const y = height - 1 - row;
      vWall[row][0] = !exitAt.has("0," + y);
      vWall[row][width] = !exitAt.has(width - 1 + "," + y);
    }

    const svgWidth = PAD * 2 + width * CELL;
    const svgHeight = PAD * 2 + height * CELL;
    const parts = [
      `<svg xmlns="http://www.w3.org/2000/svg" width="${svgWidth}" height="${svgHeight}" viewBox="0 0 ${svgWidth} ${svgHeight}">`,
      `<rect width="${svgWidth}" height="${svgHeight}" fill="${BACKGROUND}"/>`,
    ];

    const bridgeLines = [];
    for (let row = 0; row < height; row++) {
      for (let x = 0; x < width; x++) {
        const cell = grid[row][x];
        if (!cell.bridge) continue;
        addBridge(parts, bridgeLines, x, row, cell.vertical);
      }
    }

    parts.push('<g fill="none" stroke="#000" stroke-width="2" stroke-linecap="square">');
    for (let row = 0; row <= height; row++) {
      let x = 0;
      while (x < width) {
        if (!hWall[row][x]) {
          x++;
          continue;
        }
        let end = x;
        while (end + 1 < width && hWall[row][end + 1]) end++;
        parts.push(
          `<line x1="${px(x)}" y1="${py(row)}" x2="${px(end + 1)}" y2="${py(row)}"/>`
        );
        x = end + 1;
      }
    }
    for (let col = 0; col <= width; col++) {
      let row = 0;
      while (row < height) {
        if (!vWall[row][col]) {
          row++;
          continue;
        }
        let end = row;
        while (end + 1 < height && vWall[end + 1][col]) end++;
        parts.push(
          `<line x1="${px(col)}" y1="${py(row)}" x2="${px(col)}" y2="${py(end + 1)}"/>`
        );
        row = end + 1;
      }
    }
    for (const line of bridgeLines) parts.push(line);
    parts.push("</g>", "</svg>");
    return parts.join("");
  }

  // Vertical bridge geometry, in local cell coordinates, matches bridge.svg.
  // Horizontal bridges reuse it rotated 90 degrees clockwise around the cell.
  function addBridge(parts, lines, col, row, vertical) {
    const x0 = px(col);
    const y0 = py(row);
    const u = BRIDGE_UNIT;
    const bands = [
      { y: 7 * u, h: 2 * u, opacity: 0.6 },
      { y: 2 * u, h: 2 * u, opacity: 0.6 },
      { y: 4 * u, h: 3 * u, opacity: 1 },
    ];
    for (const band of bands) {
      const rect = mapRect(vertical, x0, y0, u, band.y, 9 * u, band.h);
      parts.push(
        `<rect x="${fmt(rect.x)}" y="${fmt(rect.y)}" width="${fmt(rect.w)}" height="${fmt(rect.h)}" fill="${BRIDGE_FILL}" fill-opacity="${band.opacity}"/>`
      );
    }

    // Approach ends are two short stubs each. The long crossbar that joined them
    // in bridge.svg is omitted.
    const rails = [
      [0, 0, 0, 2 * u],
      [CELL, 0, CELL, 2 * u],
      [0, 9 * u, 0, CELL],
      [CELL, 9 * u, CELL, CELL],
      [u, 2 * u, u, 9 * u],
      [10 * u, 2 * u, 10 * u, 9 * u],
    ];
    for (const [x1, y1, x2, y2] of rails) {
      const a = mapPoint(vertical, x0, y0, x1, y1);
      const b = mapPoint(vertical, x0, y0, x2, y2);
      lines.push(
        `<line x1="${fmt(a[0])}" y1="${fmt(a[1])}" x2="${fmt(b[0])}" y2="${fmt(b[1])}"/>`
      );
    }
  }

  function mapPoint(vertical, x0, y0, lx, ly) {
    if (vertical) return [x0 + lx, y0 + ly];
    return [x0 + CELL - ly, y0 + lx];
  }

  function mapRect(vertical, x0, y0, lx, ly, w, h) {
    if (vertical) return { x: x0 + lx, y: y0 + ly, w: w, h: h };
    return { x: x0 + CELL - ly - h, y: y0 + lx, w: h, h: w };
  }

  function fmt(v) {
    return String(Math.round(v * 1000) / 1000);
  }

  function px(col) {
    return PAD + col * CELL;
  }

  function py(row) {
    return PAD + row * CELL;
  }

  function parseGrid(lines) {
    const rows = mazeRows(lines);
    if (rows.length === 0) {
      throw new Error("maze text is empty");
    }
    const width = Array.from(rows[0]).length;
    if (width === 0) {
      throw new Error("maze text is empty");
    }
    return rows.map((row, rowIndex) => {
      const chars = Array.from(row);
      if (chars.length !== width) {
        throw new Error("maze rows must be the same length");
      }
      const y = rows.length - 1 - rowIndex;
      return chars.map((ch, x) => {
        if (BRIDGES.has(ch)) return { bridge: true, vertical: ch === "╫", walls: 0 };
        if (Object.prototype.hasOwnProperty.call(WALLS, ch)) {
          return { bridge: false, walls: WALLS[ch] };
        }
        throw new Error(`unknown maze character ${JSON.stringify(ch)} at (${x},${y})`);
      });
    });
  }

  function mazeRows(lines) {
    if (typeof lines === "string") lines = lines.split(/\r?\n/);
    if (!Array.isArray(lines)) {
      throw new Error("maze text must be a list of strings");
    }
    const rows = [];
    for (const raw of lines) {
      for (const piece of String(raw).split(/\r?\n/)) {
        const line = stripAnsi(piece).trim();
        if (line === "") continue;
        if (EXIT_PAIR.test(line) && !Array.from(line).some((ch) => ch in WALLS || BRIDGES.has(ch))) {
          continue;
        }
        rows.push(line);
      }
    }
    return rows;
  }

  function stripAnsi(s) {
    return s.replace(/\u001b\[[0-9;]*m/g, "");
  }

  function normalizeEntrances(entrances, width, height) {
    if (entrances == null) return [];
    if (!Array.isArray(entrances)) {
      throw new Error("entrances must be a list of coordinates");
    }
    return entrances.map((entrance, i) => {
      let x;
      let y;
      if (Array.isArray(entrance)) {
        x = entrance[0];
        y = entrance[1];
      } else if (entrance && typeof entrance === "object") {
        x = entrance.x;
        y = entrance.y;
      } else {
        throw new Error(`entrance ${i + 1} is not a coordinate`);
      }
      if (!Number.isInteger(x) || !Number.isInteger(y)) {
        throw new Error(`entrance ${i + 1} must be integer coordinates`);
      }
      if (x < 0 || y < 0 || x >= width || y >= height) {
        throw new Error(`entrance (${x},${y}) is outside the ${width}×${height} maze`);
      }
      return { x, y };
    });
  }

  return { renderMazeSvg };
});
