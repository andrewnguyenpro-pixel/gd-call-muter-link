#pragma once
// Tiny local web server; lives in server.cpp, built without Geode's
// precompiled header so winsock2.h doesn't clash.
#include <atomic>

struct LinkState {
    std::atomic<bool> inLevel{false};
    std::atomic<bool> practice{false};
    std::atomic<float> percent{0.f};
    std::atomic<long> attempt{0};
};

extern LinkState g_link;
// returns false if the port is busy
bool startLinkServer();
