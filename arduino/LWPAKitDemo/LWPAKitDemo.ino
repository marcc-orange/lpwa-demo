/**
  Copyright (C) 2015 Orange

  This software is distributed under the terms and conditions of the 'Apache-2.0'
  license which can be found in the file 'LICENSE' in this package distribution
  or at 'http://www.apache.org/licenses/LICENSE-2.0'.
*/

#include <math.h>

bool ledState = false;

const long DELAY_LOOP = 180000;
const int PIN_BUTTON = 7;        // Grove - Button
const int PIN_LIGHTSENSOR = A5;  // Grove - Temperature Sensor
const int LEDPIN = 13;           // Onboard LED

void setup()
{
  initXbeeDebugSerials(19200);  // xbee baud rate = 19200 on Atmega side

  pinMode(PIN_BUTTON, INPUT); // Set the BUTTON as INPUT
  pinMode(LEDPIN, OUTPUT);    // Set the LED on Digital 12 as an OUTPUT

  ledBlinking(LEDPIN, 10, 500);  // Blink for ten seconds
  ledOnOff(true);       // Led ON along with command mode

  // set ATO & ATM Parameters for Nano N8
  initXbeeNanoN8();

  ledOnOff(false); // Led Off after command mode
  delay(3000);
}

void loop()
{
  // Read light measure
  int sensorValue = analogRead(PIN_LIGHTSENSOR); //read light from sensor

  // Send light measure and led state
  sendLightSensorValuetoNanoN8(sensorValue);

  // Wait max 6s until data is available
  long l_milli = millis();
  while (millis() - l_milli < 6000 && Serial.available() <= 0) { }

  // Read downlink data, if any
  while (Serial.available () > 0) {
    ledBlinking(LEDPIN, 2, 200); // blink for 2s when data is available

    switch (Serial.read()) {
      case 0: ledOnOff(false);
        break;
      case 1: ledOnOff(true);
        break;
      case 2: ledBlinking(LEDPIN, 5, 500); // blink for 5s
        break;
      case -1: // EOF
        break;
      default: // unexpected data
        ledBlinking(LEDPIN, 4, 100); // blink quickly for 2s
        break;
    }
  }

  // Wait and check for button press
  l_milli = millis();
  while (millis() - l_milli < DELAY_LOOP) {
    if (digitalRead(PIN_BUTTON) != 0) {
      ledBlinking(LEDPIN, 4, 500);
      break;
    }
  }
}

void initXbeeDebugSerials(int xbee_rate)
{
  Serial.begin(xbee_rate);       // the Bee baud rate on Software Serial Atmega
  delay(1000);
}

void initXbeeNanoN8()
{
  Serial.print("+++");          // Enter command mode
  delay(1500);
  Serial.print("ATM007=06\n");    // Baud rate 19200
  delay(1500);
  Serial.print("ATQ\n");        // Quit command mode
  delay(1500);

  // Empty any initial output of the module ?
  while (Serial.read() > 0) { }
}

void ledOnOff(boolean state)
{
  digitalWrite(LEDPIN, state ? HIGH : LOW);
  ledState = state;
}

void ledBlinking(int led, int sec, int period)
{
  long end = millis() + sec * 1000;
  while (millis() < end)
  {
    digitalWrite(LEDPIN, HIGH);
    delay(period / 2);
    digitalWrite(LEDPIN, LOW);
    delay(period / 2);
  }
  ledOnOff(ledState); // restore previous led state
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

