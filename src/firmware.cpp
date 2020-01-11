#include <Particle.h>
#include <Adafruit_SSD1306.h>
#include <vector>

const uint16_t DT = D2;
const uint16_t CLK = D3;

int32_t vclkPrev = 0;
int32_t encoderIdx = 0;

auto clicklist = std::vector<std::string>{};

Adafruit_SSD1306 tft(128, 32);
void setup() {
   tft.begin(SSD1306_SWITCHCAPVCC, 0x3C);

   tft.clearDisplay();
   tft.setTextColor(WHITE);
   tft.setTextSize(1);

   tft.print("ready");
   tft.display();

   pinMode(DT, INPUT);
   pinMode(CLK, INPUT);

   vclkPrev = digitalRead(CLK);

   // these will be fetched from server
   clicklist = {"foo", "bar", "baz", "bam"};
}

void loop() {
   const auto vclk = digitalRead(CLK);
   if(vclk != vclkPrev) {
      if(digitalRead(DT) != vclk)
         ++encoderIdx;
      else --encoderIdx;
      vclkPrev = vclk;

      auto idx = encoderIdx % clicklist.size();
      if(idx < 0) idx += clicklist.size();

      tft.clearDisplay();
      tft.println(clicklist.at(idx).c_str());
      tft.printf("%i", encoderIdx);
      tft.display();
   }
}
