// GD Call Muter Link (Geode mod, Windows)
// Reports your exact level % to GD Call Muter v10 on this PC.
// The app asks http://127.0.0.1:47321/state and gets back:
//   {"inLevel":true,"percent":45.3,"practice":false,"attempt":12}
// "attempt" goes up every time you die or restart, so the app knows to unmute.

#include <Geode/Geode.hpp>
#include <Geode/modify/PlayLayer.hpp>

#include "server.hpp"

using namespace geode::prelude;

$on_mod(Loaded) {
    if (startLinkServer()) log::info("GD Call Muter Link: listening on 127.0.0.1:47321");
    else log::error("GD Call Muter Link: port 47321 is busy");
}

class $modify(CallMuterPL, PlayLayer) {
    void postUpdate(float dt) {
        PlayLayer::postUpdate(dt);
        g_link.inLevel = true;
        g_link.practice = m_isPracticeMode;
        g_link.percent = this->getCurrentPercent();
    }

    void resetLevel() {
        g_link.attempt++;          // died or restarted
        g_link.percent = 0.f;
        PlayLayer::resetLevel();
    }

    void onQuit() {
        g_link.inLevel = false;    // left the level
        g_link.percent = 0.f;
        PlayLayer::onQuit();
    }
};
