/**
  Copyright (C) 2015 Orange

  This software is distributed under the terms and conditions of the 'Apache-2.0'
  license which can be found in the file 'LICENSE' in this package distribution
  or at 'http://www.apache.org/licenses/LICENSE-2.0'.
*/

#include <math.h>
#include <SoftwareSerial.h>

bool ledState = false;

const long DELAY_LOOP = 180000;  // Delay between two transmissions
const long DELAY_ACK  =  10000;  // Delay for ACK transmission after command
const int GROVE_LIGHT = A5;      // Grove - Light Sensor
const int GROVE_LED   = 7;       // Grove - LED light (optional, set to 13 if no Grove LED)
const int ONBOARD_LED = 13;      // Onboard LED

// Debug thought soft serial (optional)
const int DEBUG_RX = 3;
const int DEBUG_TX = 2;
SoftwareSerial debug(DEBUG_RX, DEBUG_TX);

void setup()
{
  debug.begin(19200);  // baud rate for debug serial
  Serial.begin(19200); // baud rate for NanoN8

  debug.println("Setup...");

  pinMode(GROVE_LED, OUTPUT);    // Set the Grove LED as output
  pinMode(ONBOARD_LED, OUTPUT);  // Set the LED on Digital 12 as an OUTPUT

  ledBlinking(ONBOARD_LED, 10, 500);  // Blink for ten seconds
  ledOnOff(ONBOARD_LED, true);        // Led ON along with command mode

  // set ATO & ATM Parameters for Nano N8
  initXbeeNanoN8();

  ledOnOff(ONBOARD_LED, false); // Led Off after command mode
  
  debug.println("Setup: DONE");

  delay(1000);
}

void loop()
{
  // Read light measure
  int sensorValue = analogRead(GROVE_LIGHT);
  
  debug.print("Sensor: ");
  debug.println(sensorValue);

  // Send light measure and led state
  sendLightSensorValuetoNanoN8(sensorValue);

  // Wait max 6s until data is available
  long l_milli = millis();
  while (millis() - l_milli < 6000 && Serial.available() <= 0) { }

  // Time to wait for next transmission
  int wait = DELAY_LOOP;

  // Read downlink data, if any
  while (Serial.available () > 0) {
    ledBlinking(ONBOARD_LED, 2, 200); // blink for 2s when data is available

    int c = Serial.read();
    switch (c) {
      case 0:
        debug.println("LED: Off");
        ledOnOff(GROVE_LED, false);
        wait = DELAY_ACK; // retransmit ACK now
        break;
      case 1: 
        debug.println("LED: On");
        ledOnOff(GROVE_LED, true);
        wait = DELAY_ACK; // retransmit ACK now
        break;
      case 2:
        debug.println("LED: Blink");
        ledBlinking(GROVE_LED, 5, 500); // blink for 5s
        wait = DELAY_ACK; // retransmit ACK now
        break;
      case -1: // EOF
        debug.println("EOF");
        break;
      default: // unexpected data
        debug.print("Unexpected data: ");
        debug.println(c);
        ledBlinking(ONBOARD_LED, 4, 100); // blink quickly for 2s
        break;
    }
  }

  // Wait until timeout or low light condition
  l_milli = millis();
  while (millis() - l_milli < wait) {
    if (analogRead(GROVE_LIGHT) < 240) {
      ledBlinking(ONBOARD_LED, 4, 500);
      break;
    }
  }
}

void debugNanoN8(char * message, char * command)
{
  Serial.println(command);
  debug.print(message);
  debug.println(Serial.readString());
}

void initXbeeNanoN8()
{
  Serial.print("+++");          // Enter command mode
  delay(1500);
  Serial.println("ATM007=06");  // Baud rate 19200
  delay(1500);
  
  debugNanoN8("Version: ", "ATV");
  debugNanoN8("AppSKey: ", "ATO083");
  debugNanoN8("Device address: ", "ATO069");
  debugNanoN8("Network Session Key: ", "ATO073");
  debugNanoN8("Application Session Key: ", "ATO074");
  
  Serial.println("ATQ");        // Quit command mode
  delay(1500);

  // Empty any initial output of the module ?
  while (Serial.read() > 0) { }
}

void ledOnOff(int led, boolean state)
{
  digitalWrite(led, state ? HIGH : LOW);
  // For Grove LED, keep state
  if (led == GROVE_LED) {
    ledState = state;
  }
}

void ledBlinking(int led, int sec, int period)
{
  long end = millis() + sec * 1000;
  while (millis() < end)
  {
    digitalWrite(led, HIGH);
    delay(period / 2);
    digitalWrite(led, LOW);
    delay(period / 2);
  }
  // restore previous Grove LED state
  if (led == GROVE_LED) {
    ledOnOff(led, ledState); 
  }
}

void sendLightSensorValuetoNanoN8(int value)
{
  byte bLedState = ledState ? 0x01 : 0x00;
  byte byteOne = value / 256;
  byte byteTwo = value % 256;

  Serial.write(bLedState);
  Serial.write((byte)0x00);
  Serial.write((byte)0x00);
  Serial.write(byteOne);
  Serial.write(byteTwo);
}

