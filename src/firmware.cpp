#include <Particle.h>
#include <Adafruit_SSD1306.h>


Adafruit_SSD1306 tft(128, 32);
void setup() {
   tft.begin(SSD1306_SWITCHCAPVCC, 0x3C);

   tft.setTextColor(WHITE);
   tft.setTextSize(1);

   tft.print("ready");
   tft.display();
}

void loop() {
}
