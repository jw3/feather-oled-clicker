#include <Particle.h>
#include <Adafruit_SSD1306.h>
#include <Debounce.h>

#include <vector>
#include <functional>

SYSTEM_MODE(MANUAL)

const uint16_t DataPin = D6;
const uint16_t ClockPin = D8;
const uint16_t SwitchPin = D7;

int32_t vclkPrev = 0;
int32_t menuIdx = 0;
int32_t encoderIdx = 0;

SerialLogHandler logHandler;
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

   pinMode(DataPin, INPUT);
   pinMode(ClockPin, INPUT);

   pinMode(SwitchPin, INPUT);
   clicker.attach(SwitchPin);
   clicker.interval(20);

   vclkPrev = digitalRead(ClockPin);
}

long last = 0;
std::array<char, 1024> buffer = {};
bool ackd = false;
bool initd = false;
long last_click = 0;
void loop() {
   while(Serial.available()) {
      auto c = Serial.read();
      if(c != '\n') buffer[last++] = c;
      else if(last) {
         std::string s(buffer.data(), last);

//         tft.printf("%s", s.c_str());
//         tft.display();

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
               Log.info("adding model");
            }
         }
//         else {
//            Serial.printlnf("data[%i]: %s", s.length(), s.c_str());
//         }
      }
   }

   if(initd) {
      clicker.update();

      // clicker -------------------
      auto t = millis();
      if(clicker.rose()) {
         if(t - last_click < 750) {
            // double click
            auto idx = encoderIdx / 2 % clicklist.size();
            if(idx < 0) idx += clicklist.size();
            Serial.printlnf("X=%i", idx);
            tft.print("!!");
            tft.display();
         }
         last_click = t;
      }
      // ---------------------------

      const auto vclk = digitalRead(ClockPin);
      if(vclk != vclkPrev) {
         if(digitalRead(DataPin) != vclk)
            ++encoderIdx;
         else --encoderIdx;

         if(!(encoderIdx % 2)) {
            menuIdx = encoderIdx / 2 % clicklist.size();
            if(menuIdx< 0) menuIdx += clicklist.size();

            tft.clearDisplay();
            tft.setCursor(0, 0);
            tft.printf("%2i. %s", menuIdx + 1, clicklist.at(menuIdx).c_str());
            tft.display();
         }
      }
      vclkPrev = vclk;
   }
}
