#include <Particle.h>
#include <Adafruit_SSD1306.h>
#include <Debounce.h>

#include <vector>
#include <functional>

SYSTEM_MODE(MANUAL)

const uint8_t ClickedIcon[] = {0x00, 0x00, 0x00, 0x1f, 0xf8, 0x00, 0x3f, 0xf8, 0x80, 0x60, 0x01, 0x80, 0x60, 0x03, 0x00, 0x60, 0x06, 0x00, 0x60, 0x0c, 0x00, 0x62, 0x18, 0x00, 0x63, 0x31, 0x80,
                               0x61, 0xe1, 0x80, 0x60, 0xc1, 0x80, 0x60, 0x01, 0x80, 0x60, 0x01, 0x80, 0x60, 0x01, 0x80, 0x60, 0x01, 0x80, 0x3f, 0xff, 0x00, 0x1f, 0xfe, 0x00, 0x00, 0x00, 0x00};

const uint16_t DataPin = D6;
const uint16_t ClockPin = D8;
const uint16_t SwitchPin = D7;

uint16_t state = 0;
uint16_t menuIdx = 0;
uint16_t encoderIdx = 0;

std::vector<std::string> clicklist;

Debounce clicker = Debounce();

Adafruit_SSD1306 tft(128, 32);
void setup() {
   Serial.begin(9600);

   tft.begin(SSD1306_SWITCHCAPVCC, 0x3C);

   tft.clearDisplay();
   tft.setTextColor(WHITE);
   tft.setTextSize(2);
   tft.setCursor(0, 0);
   tft.print("connecting");
   tft.display();

   pinMode(DataPin, INPUT_PULLUP);
   pinMode(ClockPin, INPUT);

   pinMode(SwitchPin, INPUT);
   clicker.attach(SwitchPin);
   clicker.interval(20);
}

long last = 0;
std::array<char, 1024> buffer = {};
bool ackd = false;
bool initd = false;
long last_click = 0;
long last_dblclick = 0;
long last_dblclick_idx = -1;

void loop() {
   while(Serial.available()) {
      auto c = Serial.read();
      if(c != '\n') buffer[last++] = c;
      else if(last) {
         std::string s(buffer.data(), last);
         buffer = {};
         last = 0;

         if(!ackd && s == "hello") {
            Serial.println("HELLO");
            ackd = true;
         }
         else if(ackd && !initd) {
            if(s == "READY") {
               tft.clearDisplay();
               tft.setCursor(0, 0);
               if(clicklist.empty()) tft.println("No models");
               else tft.printlnf(" 1. %s", clicklist.front().c_str());
               tft.display();

               initd = true;

            }
            else {
               clicklist.push_back(s);
               tft.print(".");
               tft.display();
            }
         }
      }
   }

   if(initd) {
      // encoder -------------------
      state = (state << 1) | digitalRead(ClockPin) | 0xe000;
      if(state == 0xf000) {
         if(digitalRead(DataPin))
            ++encoderIdx;
         else --encoderIdx;

         state = 0;
         menuIdx = encoderIdx % clicklist.size();

         tft.clearDisplay();
         tft.setCursor(0, 0);
         tft.printf("%2i. %s", menuIdx + 1, clicklist.at(menuIdx).c_str());
         if(menuIdx == last_dblclick_idx)
            tft.drawBitmap(110, 15, ClickedIcon, 18, 18, 1);
         tft.display();
      }
      // encoder -------------------


      // clicker -------------------
      clicker.update();
      auto t = millis();
      if(clicker.rose()) {
         if(t - last_click < 750 && t - last_dblclick > 2000) {
            // double click
            Serial.printlnf("X=%i", menuIdx);
            last_dblclick_idx = menuIdx;
            tft.drawBitmap(110, 15, ClickedIcon, 18, 18, 1);
            tft.display();

            last_dblclick = t;
         }
         last_click = t;
      }
      // ---------------------------
   }
}
