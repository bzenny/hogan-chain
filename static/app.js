document.body.addEventListener("notify",(event)=>{const d=event.detail||{};console.log(`[${d.type||"info"}] ${d.message||""}`)});
