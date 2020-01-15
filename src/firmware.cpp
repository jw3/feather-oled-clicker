#include <Particle.h>
#include <Adafruit_SSD1306.h>
#include <vector>

SYSTEM_MODE(MANUAL)

const uint16_t DataPin = D6;
const uint16_t ClockPin = D8;
const uint16_t SwitchPin = D7;

int32_t vclkPrev = 0;
int32_t encoderIdx = 0;

SerialLogHandler logHandler;
auto clicklist = std::vector<std::string>{};

Adafruit_SSD1306 tft(128, 32);
void setup() {
   Serial.begin(9600);

   tft.begin(SSD1306_SWITCHCAPVCC, 0x3C);


   tft.clearDisplay();
   tft.setTextColor(WHITE);
   tft.setTextSize(1);
   tft.setCursor(0, 0);
   tft.print("initializing");
   tft.display();

   pinMode(DataPin, INPUT);
   pinMode(ClockPin, INPUT);
   pinMode(SwitchPin, INPUT);

   vclkPrev = digitalRead(ClockPin);

   // these will be fetched from server
   clicklist = {"foo", "bar", "baz", "bam"};
}

long last = 0;
std::array<char, 1024> buffer = {};
bool ackd = false;
void loop() {
   while(Serial.available()) {
      auto c = Serial.read();
      if(c != '\n') buffer[last++] = c;
      else if(last) {
         std::string s(buffer.data(), last);

         tft.printlnf("%s", s.c_str());
         tft.display();

         buffer = {};
         last = 0;

         if(!ackd && s == "hello") {
            Serial.println("HELLO");
            ackd = true;
         }
//         else {
//            Serial.printlnf("data[%i]: %s", s.length(), s.c_str());
//         }
      }
   }

   if(digitalRead(SwitchPin) == LOW) {
      Log.info("CLICKED!");
      tft.print("!");
      tft.display();

      auto idx = encoderIdx % clicklist.size();
      if(idx < 0) idx += clicklist.size();
      Serial.printlnf("X=%i", idx);
   }

   const auto vclk = digitalRead(ClockPin);
   if(vclk != vclkPrev) {
      if(digitalRead(DataPin) != vclk)
         ++encoderIdx;
      else --encoderIdx;

      auto idx = encoderIdx % clicklist.size();
      if(idx < 0) idx += clicklist.size();

      tft.clearDisplay();
      tft.setCursor(0, 0);
      tft.print(clicklist.at(idx).c_str());
      tft.printf(" %i", encoderIdx);
      Log.info("turned to index: %li", encoderIdx);
      tft.display();
   }
   vclkPrev = vclk;
}
