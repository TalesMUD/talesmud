import {
  beginWsConnect,
  registerWs,
  clearWs,
  wsBusy,
  wsTakenOver,
  markWsTakenOver,
  closeLive,
  WS_CLOSE_SESSION_REPLACED,
} from "./websocketGate.js";

function fakeSocket(state) {
  return { readyState: state, close() { this.readyState = 3; } };
}

if (!beginWsConnect()) throw new Error("first begin should succeed");
if (beginWsConnect()) throw new Error("inflight should block second begin");
const a = fakeSocket(0);
registerWs(a);
if (!wsBusy()) throw new Error("CONNECTING is busy");
if (beginWsConnect()) throw new Error("busy socket should block begin");
a.readyState = 1;
if (beginWsConnect()) throw new Error("OPEN socket should block begin");
a.readyState = 2;
if (!wsBusy()) throw new Error("CLOSING is busy");
if (beginWsConnect()) throw new Error("CLOSING socket should block begin");
a.readyState = 3;
clearWs(a);
if (wsBusy()) throw new Error("cleared socket should not be busy");
if (!beginWsConnect()) throw new Error("after clear, begin should work");
registerWs(fakeSocket(1));
markWsTakenOver();
if (!wsTakenOver()) throw new Error("taken over flag");
if (beginWsConnect()) throw new Error("4001 must never begin another connect");
closeLive();
if (beginWsConnect()) throw new Error("taken over survives closeLive");
if (WS_CLOSE_SESSION_REPLACED !== 4001) throw new Error("close code");
console.log("websocketGate ok");
