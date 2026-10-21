const room = {id:'layout-room',name:'Layout room',description:'A quiet room for responsive checks.',meta:{background:'layout-room'},exits:['north'],actions:[]};
const npcs = [
  {id:'layout-guide',displayName:'Training guide',hasDialog:true,level:1},
  {id:'layout-merchant',displayName:'Training merchant',isMerchant:true,level:1},
];
const character = {id:'layout-player',name:'Layout guest',class:'warrior',race:'human',level:1,currentHitPoints:30,maxHitPoints:30,currentMana:10,maxMana:10,gold:15};
const picture = '<svg xmlns="http://www.w3.org/2000/svg" width="512" height="512"><rect width="512" height="512" fill="#152432"/><path d="M0 512L256 100L512 512" fill="#304457"/></svg>';

export function fixtureHttp(request,response) {
  const pathname=new URL(request.url,'http://localhost').pathname;
  if(!pathname.startsWith('/api/')) return false;
  if(/backgrounds|portraits|item-art|skill-art/.test(pathname)) {
    response.writeHead(200,{'content-type':'image/svg+xml'});response.end(picture);return true;
  }
  const bodies={
    '/api/guest':{token:'local-layout-fixture',expiresIn:1800},
    '/api/server-info':{serverName:'Layout fixture',...(process.env.UX_DEFAULT_THEME?{defaultTheme:process.env.UX_DEFAULT_THEME}:{})},
    '/api/classes':{source:'pack',races:[{id:'human',name:'Human'}],classes:[{id:'warrior',name:'Warrior',templateId:'layout-template',races:['human'],skills:[],hotbarCap:4}]},
  };
  response.writeHead(200,{'content-type':'application/json'});
  response.end(JSON.stringify(bodies[pathname] || {}));
  return true;
}

export function fixtureSocket(socket) {
  const send=message=>{ if(socket.readyState===1) socket.send(JSON.stringify(message)); };
  const timer=setTimeout(()=>{
    send({type:'characterSelected',character,message:'Layout guest selected',xpForNextLevel:45});
    send({type:'enterRoom',room,npcs,players:[{...character,isYou:true}],message:'Layout room'});
    send({type:'characterUpdate',...character,equippedSkills:[],maxSkillSlots:4,attributes:[]});
    send({type:'inventoryUpdate',inventory:{items:[]},equippedItems:{},gold:15});
    send({type:'guideHint',text:'Try the training guide',command:'look',label:'Look'});
  },50);
  socket.on('close',()=>clearTimeout(timer));
  socket.on('message',data=>{
    const frame=JSON.parse(String(data));
    const command=String(frame.message || '');
    if(command.startsWith('speak to ')) send({type:'dialog',npcName:'Training guide',npcText:'Choose a training topic.',options:Array.from({length:6},(_,i)=>({index:i+1,text:i===5?'Goodbye':`Training topic ${i+1}`}))});
    else if(/^\d+$/.test(command)) send({type:'dialogEnd',message:'Conversation ended'});
    else if(command==='list') send({type:'shop',merchantName:'Training merchant',merchantId:'layout-merchant',gold:15,stock:Array.from({length:4},(_,i)=>({id:`layout-stock-${i}`,name:`Training item ${i+1}`,price:2,quantity:3,type:'misc'}))});
  });
}
