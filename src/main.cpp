// GD Call Muter Link (Geode mod, Windows)
// Reports your exact level % to GD Call Muter v10 on this PC.
// The app asks http://127.0.0.1:47321/state and gets back:
//   {"inLevel":true,"percent":45.3,"practice":false,"attempt":12}
// "attempt" goes up every time you die or restart, so the app knows to unmute.

#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>

#include <Geode/Geode.hpp>
#include <Geode/modify/PlayLayer.hpp>

#include <atomic>
#include <cstdio>
#include <string>
#include <thread>

using namespace geode::prelude;

static std::atomic<bool> g_inLevel{false};
static std::atomic<bool> g_practice{false};
static std::atomic<float> g_percent{0.f};
static std::atomic<long> g_attempt{0};

static void handleClient(SOCKET c) {
    char buf[512];
    recv(c, buf, sizeof(buf), 0);
    char body[160];
    std::snprintf(body, sizeof(body),
        "{\"inLevel\":%s,\"percent\":%.3f,\"practice\":%s,\"attempt\":%ld}",
        g_inLevel ? "true" : "false", (double)g_percent.load(),
        g_practice ? "true" : "false", g_attempt.load());
    std::string resp =
        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n"
        "Access-Control-Allow-Origin: *\r\nCache-Control: no-store\r\n"
        "Content-Length: " + std::to_string(std::strlen(body)) +
        "\r\nConnection: close\r\n\r\n" + body;
    send(c, resp.c_str(), (int)resp.size(), 0);
    closesocket(c);
}

static void serverLoop() {
    WSADATA wd;
    WSAStartup(MAKEWORD(2, 2), &wd);
    SOCKET s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(47321);
    addr.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    if (bind(s, (sockaddr*)&addr, sizeof(addr)) != 0 || listen(s, 16) != 0) {
        log::error("GD Call Muter Link: port 47321 is busy");
        return;
    }
    log::info("GD Call Muter Link: listening on 127.0.0.1:47321");
    while (true) {
        SOCKET c = accept(s, nullptr, nullptr);
        if (c == INVALID_SOCKET) continue;
        handleClient(c);   // each reply is tiny, so one at a time is fine
    }
}

$on_mod(Loaded) {
    std::thread(serverLoop).detach();
}

class $modify(CallMuterPL, PlayLayer) {
    void postUpdate(float dt) {
        PlayLayer::postUpdate(dt);
        g_inLevel = true;
        g_practice = m_isPracticeMode;
        g_percent = this->getCurrentPercent();
    }

    void resetLevel() {
        g_attempt++;          // died or restarted
        g_percent = 0.f;
        PlayLayer::resetLevel();
    }

    void onQuit() {
        g_inLevel = false;    // left the level
        g_percent = 0.f;
        PlayLayer::onQuit();
    }
};
