import { hash } from './mapArt.js';

// Trace the union of courtyards and charted streets on a quarter-cell grid.
// Straight boundary segments retain real corners and concave town footprints.
export function townFortifications(town, roads) {
  const occupied=new Set(), ids=new Set(town.map(p=>p.id));
  const add=(x,y,r)=>{
    for(let gx=Math.floor((x-r)*4);gx<Math.ceil((x+r)*4);gx++)
      for(let gy=Math.floor((y-r)*4);gy<Math.ceil((y+r)*4);gy++)occupied.add(`${gx}:${gy}`);
  };
  for(const p of town)add(p.x,p.y,.72);
  for(const {a,b} of roads)if(ids.has(a.id)&&ids.has(b.id)) {
    const steps=Math.ceil(Math.hypot(a.x-b.x,a.y-b.y)*8);
    for(let i=0;i<=steps;i++)add(a.x+(b.x-a.x)*i/Math.max(1,steps),a.y+(b.y-a.y)*i/Math.max(1,steps),.42);
  }
  const edges=new Map();
  const edge=(x,y,xx,yy)=>edges.set(`${x}:${y}`,{x:xx,y:yy});
  for(const key of occupied) {
    const [x,y]=key.split(':').map(Number);
    if(!occupied.has(`${x}:${y-1}`))edge(x,y,x+1,y);
    if(!occupied.has(`${x+1}:${y}`))edge(x+1,y,x+1,y+1);
    if(!occupied.has(`${x}:${y+1}`))edge(x+1,y+1,x,y+1);
    if(!occupied.has(`${x-1}:${y}`))edge(x,y+1,x,y);
  }
  const loops=[];
  while(edges.size) {
    const start=edges.keys().next().value,loop=[];let key=start;
    while(edges.has(key)) {
      const [x,y]=key.split(':').map(Number);loop.push({x:x/4,y:y/4});
      const end=edges.get(key);edges.delete(key);key=`${end.x}:${end.y}`;
      if(key===start)break;
    }
    const corners=loop.filter((p,i)=>{
      const a=loop[(i+loop.length-1)%loop.length],b=loop[(i+1)%loop.length];
      return (p.x-a.x)*(b.y-p.y)!==(p.y-a.y)*(b.x-p.x);
    });
    if(corners.length>=4)loops.push(corners);
  }
  const gates=[];
  for(const road of roads) {
    if(ids.has(road.a.id)===ids.has(road.b.id))continue;
    const a=ids.has(road.a.id)?road.a:road.b,b=a===road.a?road.b:road.a;
    const dx=b.x-a.x,dy=b.y-a.y;let crossing=null;
    for(const loop of loops)for(let i=0;i<loop.length;i++) {
      const p=loop[i],q=loop[(i+1)%loop.length],ex=q.x-p.x,ey=q.y-p.y;
      const denom=dx*ey-dy*ex;if(!denom)continue;
      const t=((p.x-a.x)*ey-(p.y-a.y)*ex)/denom,u=((p.x-a.x)*dy-(p.y-a.y)*dx)/denom;
      if(t>=0&&t<=1&&u>=0&&u<=1&&(!crossing||t<crossing.t))crossing={x:a.x+dx*t,y:a.y+dy*t,t,angle:Math.atan2(dy,dx),roomId:a.id};
    }
    if(crossing&&!gates.some(g=>Math.hypot(g.x-crossing.x,g.y-crossing.y)<.5))gates.push(crossing);
  }
  const towers=[];
  for(const loop of loops)for(let i=0;i<loop.length;i++) {
    const p=loop[i],a=loop[(i+loop.length-1)%loop.length],b=loop[(i+1)%loop.length];
    const convex=(p.x-a.x)*(b.y-p.y)-(p.y-a.y)*(b.x-p.x)>0;
    if(convex&&!towers.some(t=>Math.hypot(t.x-p.x,t.y-p.y)<.9)&&!gates.some(g=>Math.hypot(g.x-p.x,g.y-p.y)<.55))towers.push(p);
  }
  return {loops,gates,towers};
}

export function mountainDepth(cells,byCell) {
  const mountain=c=>c&&['mountain','snow'].includes(c.terrain),depth=new Map(),queue=[];
  for(const c of cells)if(mountain(c)&&[[1,0],[-1,0],[0,1],[0,-1]].some(([dx,dy])=>!mountain(byCell.get(`${c.x+dx}:${c.y+dy}`)))) {
    depth.set(`${c.x}:${c.y}`,0);queue.push(c);
  }
  for(let i=0;i<queue.length;i++) {
    const c=queue[i],d=depth.get(`${c.x}:${c.y}`);
    for(const [dx,dy] of [[1,0],[-1,0],[0,1],[0,-1]]) {
      const key=`${c.x+dx}:${c.y+dy}`,n=byCell.get(key);
      if(mountain(n)&&!depth.has(key)){depth.set(key,d+1);queue.push(n)}
    }
  }
  return depth;
}
export function reliefAt(cell,depth) {
  const n=hash(`${cell.x}:${cell.y}`),d=depth.get(`${cell.x}:${cell.y}`)||0;
  if(d<2)return {kind:'hill',variant:n%2,size:56+n%12};
  return {kind:'ridge',variant:d<4?2+n%2:4+n%2,size:76+Math.min(d,5)*4};
}
export function forestSpecies(area,terrain) {
  if(terrain==='swamp'||/marsh|bog|fen/i.test(area))return 'deadTree';
  if(/highland|foothill|mountain|alpine/i.test(area))return 'pine';
  return 'oak';
}
