// JQuery ;)
var $ = function(a) { return document.getElementById(a); }

var chartCtx = $('chart');
var buttonOn = $('light-on');
var buttonOff = $('light-off');
var buttonBlink = $('light-blink');
var deviceLog = $('device-log');
var commandLog = $('command-log');

// Chart setup
var labels = [];
var luminosityValues = [];
var lightValues = [];
var chartData = {
    labels:labels,
    datasets: [
        {
            label: "Luminosity",
            scaleBeginAtZero: true,
            borderColor: "rgba(179,181,198,1)",
            data: luminosityValues
        },
        {
            label: "Light state",
            scaleBeginAtZero: true,
            borderColor: "rgba(255,99,132,1)",
            lineTension: 0,
            data: lightValues
        }
    ]
};
var options = {
    responsive: true,
    maintainAspectRatio: true,
    scales: {
        yAxes: [{
            ticks: {
                beginAtZero:true
            }
        }]
    }
}
var chart = Chart.Line(chartCtx, {data: chartData, options: options });

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
              var date = new Date().toTimeString().split(" ")[0]

              if (e.type == "device") {
                  var message = 'Light: ' + e.device.lightOn + ' Luminosity: ' + e.device.luminosity + '<br />';

                  labels.push(date)
                  luminosityValues.push(e.device.luminosity)
                  lightValues.push(e.device.lightOn ? 500 : 0)
                  chart.update()

                  deviceLog.innerHTML = date + ": " + message + deviceLog.innerHTML;
                  deviceLog.scrollTop = 0;
              } else if (e.type == "command") {
                  var message = 'Command: ' + e.command.fCnt + ' State: ' + e.command.state + '<br />';
                  commandLog.innerHTML = date + ": " + message + commandLog.innerHTML;
                  commandLog.scrollTop = 0;
              }
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
