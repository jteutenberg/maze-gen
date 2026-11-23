"use strict";

const assert = require("assert");
const fs = require("fs");
const path = require("path");
const { renderMazeSvg } = require("./maze.js");

const CELL = 32;
const PAD = 2;

const sample = fs.readFileSync(path.join(__dirname, "..", "output.txt"), "utf8");
const sampleLines = sample.split(/\r?\n/);

function linesOf(svg) {
  const found = [];
  const re = /<line x1="([\d.]+)" y1="([\d.]+)" x2="([\d.]+)" y2="([\d.]+)"/g;
  let match;
  while ((match = re.exec(svg))) {
    found.push(match.slice(1).map(Number));
  }
  return found;
}

function onVertical(segments, x, y) {
  return segments.some(
    ([x1, y1, x2, y2]) => x1 === x && x2 === x && Math.min(y1, y2) <= y && y <= Math.max(y1, y2)
  );
}

function bridgeDecks(svg) {
  const found = [];
  const re = /<rect\b([^>]*)>/g;
  let match;
  while ((match = re.exec(svg))) {
    const attrs = match[1];
    const get = (name) => {
      const attr = new RegExp(name + '="([^"]*)"').exec(attrs);
      return attr ? attr[1] : null;
    };
    if (get("fill") !== "#feffff" || get("fill-opacity") !== "1") continue;
    found.push({
      x: Number(get("x")),
      y: Number(get("y")),
      width: Number(get("width")),
      height: Number(get("height")),
    });
  }
  return found;
}

function onHorizontal(segments, x, y) {
  return segments.some(
    ([x1, y1, x2, y2]) => y1 === y && y2 === y && Math.min(x1, x2) <= x && x <= Math.max(x1, x2)
  );
}

const svg = renderMazeSvg(sampleLines, [
  { x: 6, y: 1 },
  { x: 0, y: 3 },
]);

assert.match(svg, /<rect width="228" height="228" fill="#d3d3d3"\/>/);
assert.match(svg, /stroke="#000" stroke-width="2"/);
assert.equal(bridgeDecks(svg).length, 1, "sample has one bridge deck");
const sampleDeck = bridgeDecks(svg)[0];
assert.ok(sampleDeck.width > sampleDeck.height, "vertical bridge deck is a horizontal band");
assert.ok(sampleDeck.x > 194 && sampleDeck.x < 226, "deck sits inside the bridge cell");
assert.ok(sampleDeck.y > 162 && sampleDeck.y < 194, "deck sits inside the bridge cell");
assert.equal((svg.match(/fill-opacity="0.6"/g) || []).length, 2, "two translucent shoulders");

const segments = linesOf(svg);
assert.ok(onVertical(segments, PAD, PAD + CELL / 2), "top-left outer wall");
assert.ok(!onVertical(segments, PAD, PAD + 3 * CELL + CELL / 2), "west entrance is open");
assert.ok(onVertical(segments, PAD, PAD + 2 * CELL + CELL / 2), "wall above the west entrance");
assert.ok(!onVertical(segments, PAD + 7 * CELL, PAD + 5 * CELL + CELL / 2), "east entrance is open");
assert.ok(onVertical(segments, PAD + CELL, PAD + CELL / 2), "disconnected top cells");
assert.ok(!onVertical(segments, PAD + 3 * CELL, PAD + CELL / 2), "connected top corridor");
assert.ok(onHorizontal(segments, PAD + CELL / 2, PAD), "north boundary");
assert.ok(onHorizontal(segments, PAD + CELL / 2, PAD + 7 * CELL), "south boundary");

const open = renderMazeSvg(["╬╬", "╬╬"], []);
const openSegments = linesOf(open);
assert.equal(openSegments.length, 4, "outer boundary only");
assert.ok(!onVertical(openSegments, PAD + CELL, PAD + CELL / 2));

const corner = renderMazeSvg(["╔╗", "╚╝"], [[0, 0]]);
const cornerSegments = linesOf(corner);
assert.ok(!onVertical(cornerSegments, PAD, PAD + CELL + CELL / 2), "corner west side is open");
assert.ok(!onHorizontal(cornerSegments, PAD + CELL / 2, PAD + 2 * CELL), "corner south side is open");
assert.ok(onVertical(cornerSegments, PAD, PAD + CELL / 2), "west wall above the corner stays");
assert.ok(onHorizontal(cornerSegments, PAD + CELL + CELL / 2, PAD + 2 * CELL), "south wall beside the corner stays");
assert.ok(onHorizontal(cornerSegments, PAD + CELL / 2, PAD), "north boundary stays");

const bridged = renderMazeSvg(["╪"], []);
const bridgedLines = linesOf(bridged);
assert.equal(bridgeDecks(bridged).length, 1);
assert.ok(bridgeDecks(bridged)[0].height > bridgeDecks(bridged)[0].width, "horizontal bridge is rotated");
assert.ok(onVertical(bridgedLines, PAD, PAD + CELL / 2), "bridge still has an outer wall");
assert.ok(onHorizontal(bridgedLines, PAD + CELL / 2, PAD), "bridge still has an outer wall");
assert.doesNotMatch(bridged, /stroke="#aaccff"/);
assert.doesNotMatch(bridged, /width="32" height="32" fill="#ffffff"/);

const vertical = renderMazeSvg(["╫"], []);
assert.ok(bridgeDecks(vertical)[0].width > bridgeDecks(vertical)[0].height);

assert.throws(() => renderMazeSvg(sampleLines, [{ x: 9, y: 1 }]), /outside/);
assert.throws(() => renderMazeSvg(["╔╔", "╗"], []), /same length/);

const ansi = renderMazeSvg(["\u001b[1;33m╔╗\u001b[0m", "╚╝"], []);
assert.match(ansi, /<svg /);

console.log("ok");
