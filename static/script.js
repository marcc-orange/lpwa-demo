// JQuery ;)
var $ = function(a) { return document.getElementById(a); }

var buttonOn = $('light-on');
var buttonOff = $('light-off');
var buttonBlink = $('light-blink');
var contentMessage = $('echo-content');

var WebsocketClass = function(host){
    this.socket = new WebSocket(host);
    this.console = $('console');
};

WebsocketClass.prototype = {
       initWebsocket : function(){
           var $this = this;
           this.socket.onopen = function(){
               $this.log('socket opened:'+ this.readyState);
               buttonOff.disabled = false;
               buttonOn.disabled = false;
               buttonBlink.disabled = false;
           };
           this.socket.onmessage = function(e){
              $this.log('message received');
              var e = JSON.parse(e.data);
              var time = new Date();
              var message = time.getHours() + ":" + time.getMinutes() + ":" + time.getSeconds() + ": ";
              message += 'Light: ' + e.lightOn + ' Luminosity: ' + e.luminosity + '<br />';
              contentMessage.innerHTML = message + contentMessage.innerHTML;
              contentMessage.scrollTop = 0;
          };
          this.socket.onclose = function(){
               $this.log('socket closed');
               buttonOff.disabled = true;
               buttonOn.disabled = true;
               buttonBlink.disabled = true;
           };
           this.socket.onerror = function(error){
               $this.log('socket error:'+ error)
           };
           this.log('websocket init');
       },
       log: function(m) {
           console.log(m);
           this.console.innerHTML = m + '<br />' + this.console.innerHTML;
           contentMessage.scrollTop = 0;
       },
       sendMessage: function(lightStatus){
           var message = '{"light":"' + lightStatus + '"}';
           this.socket.send(message);
           this.log('message sent')
       }
};

var socket = new WebsocketClass('ws://lwpa-dev.kermit.orange-labs.fr/_ws/demo/ws');
socket.initWebsocket();

if (buttonOn.addEventListener) {
        buttonOn.addEventListener('click',function(e){
           e.preventDefault();
           socket.sendMessage("on");
           return false;
       }, true);
       buttonOff.addEventListener('click',function(e){
           e.preventDefault();
           socket.sendMessage("off");
           return false;
       }, true);
       buttonBlink.addEventListener('click',function(e){
           e.preventDefault();
           socket.sendMessage("blink");
           return false;
       }, true);
} else {
       console.log('your browser does not support addevenlistener');
}
