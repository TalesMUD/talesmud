// Move a node to <body>. HUD widgets sit in transformed / backdrop-filtered
// grid items, which turn position:fixed into "fixed to the widget" and trap
// z-index. Dialogs and their scrims must live on the page layer.
export function portal(node) {
  if (typeof document === 'undefined' || !document.body) return {};
  document.body.appendChild(node);
  return {
    destroy() {
      if (node.parentNode) node.parentNode.removeChild(node);
    },
  };
}
