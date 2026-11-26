// Tiny local web server for GD Call Muter. Built without Geode's precompiled
// header so the OS socket headers don't clash with it. Works on Windows and Mac.
#ifdef _WIN32
#define WIN32_LEAN_AND_MEAN
#include <winsock2.h>
using sock_t = SOCKET;
static void closeSock(sock_t s) { closesocket(s); }
static bool badSock(sock_t s) { return s == INVALID_SOCKET; }
#else
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>
using sock_t = int;
static void closeSock(sock_t s) { close(s); }
static bool badSock(sock_t s) { return s < 0; }
#endif

#include "server.hpp"

#include <cstdio>
#include <cstring>
#include <string>
#include <thread>

LinkState g_link;

static void handleClient(sock_t c) {
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
    closeSock(c);
}

bool startLinkServer() {
#ifdef _WIN32
    WSADATA wd;
    WSAStartup(MAKEWORD(2, 2), &wd);
#endif
    sock_t s = socket(AF_INET, SOCK_STREAM, IPPROTO_TCP);
    if (badSock(s)) return false;
    int yes = 1;
    setsockopt(s, SOL_SOCKET, SO_REUSEADDR, (const char*)&yes, sizeof(yes));
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(47321);
    addr.sin_addr.s_addr = htonl(INADDR_LOOPBACK);
    if (bind(s, (sockaddr*)&addr, sizeof(addr)) != 0 || listen(s, 16) != 0) {
        closeSock(s);
        return false;
    }
    std::thread([s] {
        while (true) {
            sock_t c = accept(s, nullptr, nullptr);
            if (badSock(c)) continue;
            handleClient(c);
        }
    }).detach();
    return true;
}
