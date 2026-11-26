#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>

#include "server.hpp"

#include <cstdio>
#include <cstring>
#include <string>
#include <thread>

LinkState g_link;

static void handleClient(SOCKET c) {
    char buf[512];
    recv(c, buf, sizeof(buf), 0);
    char body[160];
    std::snprintf(body, sizeof(body),
        "{\"inLevel\":%s,\"percent\":%.3f,\"practice\":%s,\"attempt\":%ld}",
        g_link.inLevel ? "true" : "false", (double)g_link.percent.load(),
        g_link.practice ? "true" : "false", g_link.attempt.load());
    std::string resp =
        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n"
        "Access-Control-Allow-Origin: *\r\nCache-Control: no-store\r\n"
        "Content-Length: " + std::to_string(std::strlen(body)) +
        "\r\nConnection: close\r\n\r\n" + body;
    send(c, resp.c_str(), (int)resp.size(), 0);
    closesocket(c);
}

bool startLinkServer() {
    WSADATA wd;
    WSAStartup(MAKEWORD(2, 2), &wd);
    SOCKET s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(47321);
    addr.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    if (bind(s, (sockaddr*)&addr, sizeof(addr)) != 0 || listen(s, 16) != 0) {
        closesocket(s);
        return false;
    }
    std::thread([s] {
        while (true) {
            SOCKET c = accept(s, nullptr, nullptr);
            if (c == INVALID_SOCKET) continue;
            handleClient(c);
        }
    }).detach();
    return true;
}
