import "./app.css";
import App from "./App.svelte";

Object.defineProperty(String.prototype, "hashCode", {
  value: function () {
    let hash = 0;
    for (let i = 0; i < this.length; i += 1) {
      hash = (hash << 5) - hash + this.charCodeAt(i);
      hash |= 0;
    }
    return hash;
  },
});

const app = new App({
  target: document.getElementById("app"),
});

export default app;