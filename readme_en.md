# TeleCloud

<div align="center">

[🇻🇳 Tiếng Việt](./readme.md) | 🇺🇸 English

**[📢 Support Group](https://t.me/+p-d0qfGRbX4wNzJl)**
*Join the group to discuss and get support*

</div>

**TeleCloud** is a project that allows you to use Telegram’s nearly unlimited storage capacity to store and manage files. Completely rewritten in Golang for excellent performance and low memory usage.

---

## 📸 Preview

### 🖥️ Desktop Interface
| | |
| :---: | :---: |
| <img src="preview/preview.jpg" width="100%"> | <img src="preview/preview-2.jpg" width="100%"> |
| <img src="preview/preview-3.jpg" width="100%"> | <img src="preview/preview-4.jpg" width="100%"> |

### 📱 Mobile Interface
| | | | | |
| :---: | :---: | :---: | :---: | :---: |
| <img src="preview/preview-5.jpg" width="100%"> | <img src="preview/preview-6.jpg" width="100%"> | <img src="preview/preview-7.jpg" width="100%"> | <img src="preview/preview-8.jpg" width="100%"> | <img src="preview/preview-9.jpg" width="100%"> |

---

## ✨ Features

* 📁 **Unlimited Storage**: Store files directly on Telegram with **no size limits** (Automatically splits large files into chunks from 500MB to 4GB).
* 🎬 **Media Streaming**: Stream videos and music directly in the dashboard and shared links (Seamless playback of split files).
* 🔗 **Flexible Sharing**: Supports normal or direct download links (Direct Link) for both files and **folders**.
* 🗂️ **Intuitive Management**: File Browser with **Grid** and **List** view modes.
* ⬆️ **High Performance**: Multi-threaded and chunked uploads for maximum speed and stability.
* 📂 **WebDAV Support**: Mount TeleCloud as a network drive on Windows, macOS, and Linux.
* 🔌 **Upload API**: Remote file uploads via HTTP API for script or CI/CD integration.
* 📥 **URL & Media Downloader**: Download files from URLs and Media (YouTube, TikTok, Facebook...) using **yt-dlp** directly in the UI.
* ⚡ **Background Tasks**: Background URL downloads with real-time progress notifications.
* 🧲 **Torrent Support**: Download Torrents and Magnet links directly to Telegram via **aria2c**.
* 👥 **Multi-user**: Support for child accounts with isolated storage spaces (Virtual Path).
* 🤖 **Bot Pool**: Use secondary bots to balance load, maximizing speed and reliability.
* 🔐 **Passkey Security**: Biometric login (Fingerprint, FaceID) or security keys (WebAuthn).
* 🗄️ **Multi-Database**: Supports **SQLite**, **MySQL**, and **PostgreSQL** for enterprise-scale needs.
* 🗑️ **Trash Bin**: Recover deleted files and protect data from accidental removal.
* 🔒 **Protected Shares**: Set password protection for shared files and folders.
* 🛡️ **Auto Backup**: Daily automated backups of database and thumbnails to Telegram.
* 🌐 **Multi-language**: Supports English, Vietnamese, Chinese, Japanese, Russian, and more.

---

## 🚀 Quick Start

Using the automated script is the easiest way to get started:

### Linux / Termux / macOS / Raspberry Pi
```bash
curl -fsSL https://raw.githubusercontent.com/dabeecao/telecloud-go/main/auto-setup-en.sh -o auto-setup-en.sh && bash auto-setup-en.sh
```

### Windows
Download [**`auto-install-en.bat`**](https://raw.githubusercontent.com/dabeecao/telecloud-go/main/auto-install-en.bat) and run as **Administrator**.

---

## 📖 Detailed Documentation (Wiki)

For configuration details and alternative installation methods, please refer to the documentation:

*   [🛠️ **Installation Guide**](./docs/Installation.md) (Binary, Windows, Linux...)
*   [⚙️ **Configuration Guide**](./docs/Configuration.md) (.env, Nginx Proxy...)
*   [🐳 **Docker Deployment**](./docs/Docker.md) (Docker Run, Compose)
*   [🔌 **API Documentation**](./docs/API.md) (Upload API Guide)
*   [🛠️ **Development & Localization**](./docs/Development.md) (Build from source, Contribute)

---

## 🔐 Security

TeleCloud encrypts sensitive data at rest and ships with sane hardening defaults, but the operator still has to wire a few things up.

**Required:**

*   **`TELECLOUD_MASTER_KEY`** — 32 random bytes (`openssl rand -hex 32`). Used to AES-256-GCM the Telegram session blob and sensitive settings (`api_id`, `api_hash`, `log_group_id`, `bot_tokens`). **Back this key up separately from the database.** Losing it means you have to re-authenticate Telegram and re-enter bot tokens. `auto-setup.sh` generates it into `.env`; copy it into a password manager.

**Recommended:**

*   **`TELECLOUD_SETUP_TOKEN`** — one-time token that gates `/setup` so a scanner can't race you to the admin form. Generate with `openssl rand -hex 16`. When set, requests to `/setup` must carry `X-Setup-Token` or `?token=<value>`.
*   **`LISTEN_ADDR`** — defaults to `127.0.0.1` until admin is created (so the open wizard isn't reachable from the Internet) and `0.0.0.0` afterwards. Front the service with Cloudflare Tunnel / Nginx / Tailscale instead of exposing the port directly.
*   **Docker Compose** maps to `127.0.0.1:8091` by default. Only change it to `0.0.0.0` if you deliberately want it public.

**Already on by default:**

*   Telegram sessions + sensitive settings are AES-256-GCM encrypted at rest with an auto-migration on boot (SQLite makes a `.pre-enc-<ts>.bak`; MySQL/Postgres require `TELECLOUD_I_HAVE_BACKED_UP=1` so you don't migrate without a dump).
*   `auto-setup.sh` writes a systemd unit running under a dedicated user (`User=telecloud` or `DynamicUser`), with `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp`, and SHA-256-verifies the binary against `checksums.txt` from the same release.
*   Direct-download token HMAC key is HKDF-derived from the master key, not the admin password — rotating the password does NOT invalidate previously issued links.
*   Password-protected shares mint a random session token; the bcrypt hash never leaves the server.
*   WebDAV applies a 5 attempts / 15 minutes per-IP rate limit; the auth cache uses SHA-256 of the password (not plaintext) with a 2-minute TTL, and is invalidated automatically on password change.
*   Web sessions carry `expires_at` (30 days). Changing the password invalidates every other session of that user. A background task purges expired rows every 6 hours.
*   Audit log (`audit_log`) records login/logout, password changes, admin reset, setup completion, and setting toggles. Retained 90 days.
*   Remote URL fetchers (upload-from-URL, yt-dlp, …) go through `SafeHTTPClient`, which re-resolves and pins the IP at dial time so DNS rebinding can't redirect the request to a private address.
*   CSP **excludes** Cloudflare Web Analytics by default. Flip `analytics_enabled=true` to opt in. The policy still allows `'unsafe-inline'` and `'unsafe-eval'` because the frontend uses Alpine.js.

**Known limitations:**

*   Telegram cloud chat is **not** end-to-end encrypted. Don't keep ID documents, intimate photos, secret keys or other sensitive material in TeleCloud — encrypt client-side (`rclone crypt`, Cryptomator) first if you really want to use it as storage.
*   Telegram may ban accounts used as personal cloud storage; there is no recovery. Use a dedicated phone number + account.

---

## ⚠️ Terms of Use & Disclaimer
 
**TeleCloud** is developed for storing and managing legitimate personal files. We are not responsible for any content uploaded by users or violations of Telegram’s terms of service. Users are **fully responsible** for their own actions.

The project is provided **“as-is”**, without any guarantees of stability or security.

---

## 🙏 Credits

This project uses amazing libraries:
* [gotd/td](https://github.com/gotd/td): Telegram client (MTProto API)
* [Gin](https://github.com/gin-gonic/gin): High-performance HTTP web framework
* [AlpineJS](https://github.com/alpinejs/alpine): Minimal JS framework
* [TailwindCSS](https://github.com/tailwindlabs/tailwindcss): Utility-first CSS framework
* [plyr](https://github.com/sampotts/plyr): HTML5 media player
* [Artplayer.js](https://github.com/zhw2590582/ArtPlayer): Modern and full-featured HTML5 video player.
* [Prism.js](https://github.com/PrismJS/prism): Lightweight, extensible syntax highlighter.
* [FontAwesome](https://fontawesome.com): The world's most popular icon set.
* [yt-dlp](https://github.com/yt-dlp/yt-dlp): Audio/video downloader.
* [aria2](https://github.com/aria2/aria2): Multi-protocol download utility.
* [Google Fonts (Nunito)](https://fonts.google.com/specimen/Nunito): Modern sans-serif typeface.

Thanks to all development teams and **contributors** for providing great tools and efforts for the community.

<a href="https://github.com/dabeecao/telecloud-go/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=dabeecao/telecloud-go" />
</a>

**A portion of the project's source code and this readme was referenced and modified by Gemini AI.**

---

## 📜 License

This project is licensed under the [GNU Affero General Public License v3.0 (AGPL-3.0)](https://www.gnu.org/licenses/agpl-3.0.html).
