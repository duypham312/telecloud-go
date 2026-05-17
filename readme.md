# TeleCloud

<div align="center">

🇻🇳 Tiếng Việt | [🇺🇸 English](./readme_en.md)

**[📢 Nhóm Hỗ trợ](https://t.me/+p-d0qfGRbX4wNzJl)**
*Tham gia để thảo luận và nhận hỗ trợ*

</div>

**TeleCloud** là một dự án sử dụng dung lượng lưu trữ của Telegram để lưu trữ và quản lý tệp. Được viết lại hoàn toàn bằng Golang, đem lại hiệu năng xuất sắc và sử dụng bộ nhớ cực thấp.

---

## 📸 Ảnh xem trước giao diện

### 🖥️ Giao diện Máy tính
| | |
| :---: | :---: |
| <img src="preview/preview.jpg" width="100%"> | <img src="preview/preview-2.jpg" width="100%"> |
| <img src="preview/preview-3.jpg" width="100%"> | <img src="preview/preview-4.jpg" width="100%"> |

### 📱 Giao diện Điện thoại
| | | | | |
| :---: | :---: | :---: | :---: | :---: |
| <img src="preview/preview-5.jpg" width="100%"> | <img src="preview/preview-6.jpg" width="100%"> | <img src="preview/preview-7.jpg" width="100%"> | <img src="preview/preview-8.jpg" width="100%"> | <img src="preview/preview-9.jpg" width="100%"> |

---

## ✨ Tính năng

* 📁 **Lưu trữ không giới hạn**: Lưu file trực tiếp trên Telegram **không giới hạn dung lượng** (Tự động chia nhỏ file siêu lớn thành các mảnh từ 500MB đến 4GB).
* 🎬 **Phát phương tiện**: Phát video và nhạc trực tiếp trong trang quản lý và liên kết chia sẻ (Hỗ trợ phát mượt mà các file đã chia nhỏ).
* 🔗 **Chia sẻ linh hoạt**: Hỗ trợ liên kết thường hoặc link tải trực tiếp (Direct Link), hỗ trợ chia sẻ cả **Thư mục**.
* 🗂️ **Quản lý trực quan**: Giao diện File Browser hỗ trợ chế độ xem **Lưới (Grid)** và **Danh sách (List)**.
* ⬆️ **Tốc độ tối ưu**: Upload song song (Multi-threading) và chia nhỏ (chunk) để tối ưu tốc độ và ổn định.
* 📂 **Hỗ trợ WebDAV**: Gắn TeleCloud thành ổ đĩa mạng trên máy tính (Windows, macOS, Linux).
* 🔌 **Upload API**: Cho phép upload file từ xa qua HTTP API để tích hợp vào script hoặc CI/CD.
* 📥 **Tải từ URL & Media**: Hỗ trợ tải tệp từ URL và Video/Nhạc (YouTube, TikTok, Facebook...) bằng **yt-dlp** ngay trong giao diện.
* ⚡ **Tải trong nền**: Hỗ trợ tải tệp từ URL trong nền, không cần treo trình duyệt, có thông báo tiến trình real-time.
* 🧲 **Tải Torrent**: Hỗ trợ tải Torrent và Magnet link trực tiếp về Telegram thông qua **aria2c**.
* 👥 **Đa người dùng**: Hỗ trợ tạo tài khoản con với không gian lưu trữ riêng biệt (Virtual Path).
* 🤖 **Multi-Bot (Bot Pool)**: Sử dụng nhiều Bot phụ để chia đều tải trọng, tăng tối đa tốc độ và độ ổn định.
* 🔐 **Bảo mật Passkey**: Hỗ trợ đăng nhập bằng vân tay, khuôn mặt hoặc khóa bảo mật (WebAuthn).
* 🗄️ **Đa cơ sở dữ liệu**: Hỗ trợ **SQLite**, **MySQL** và **PostgreSQL** cho các hệ thống lớn.
* 🗑️ **Thùng rác**: Lưu trữ và khôi phục các tệp đã xóa, bảo vệ dữ liệu khỏi việc xóa nhầm.
* 🔒 **Bảo mật chia sẻ**: Thiết lập mật khẩu bảo vệ cho các liên kết chia sẻ tệp và thư mục.
* 🛡️ **Sao lưu tự động**: Tự động sao lưu hàng ngày cơ sở dữ liệu và thumbnails trực tiếp vào Telegram.
* 🌐 **Đa ngôn ngữ**: Hỗ trợ Tiếng Việt, Tiếng Anh, Tiếng Trung, Tiếng Nhật, Tiếng Nga và nhiều ngôn ngữ khác.

---

## 🚀 Cài đặt nhanh

Sử dụng script tự động là cách đơn giản nhất để bắt đầu:

### Linux / Termux / macOS / Raspberry Pi
```bash
curl -fsSL https://raw.githubusercontent.com/dabeecao/telecloud-go/main/auto-setup.sh -o auto-setup.sh && bash auto-setup.sh
```

### Windows
Tải [**`auto-install.bat`**](https://raw.githubusercontent.com/dabeecao/telecloud-go/main/auto-install.bat) và chạy với quyền **Administrator**.

---

## 📖 Tài liệu chi tiết (Wiki)

Để biết thêm chi tiết về cấu hình và các phương pháp cài đặt khác, vui lòng xem tài liệu:

*   [🛠️ **Hướng dẫn cài đặt**](./docs/Installation.md) (Binary, Windows, Linux...)
*   [⚙️ **Hướng dẫn cấu hình**](./docs/Configuration.md) (.env, Nginx Proxy...)
*   [🐳 **Triển khai với Docker**](./docs/Docker.md) (Docker Run, Compose)
*   [🔌 **Tài liệu API**](./docs/API.md) (Hướng dẫn Upload API)
*   [🛠️ **Phát triển & Bản dịch**](./docs/Development.md) (Build từ nguồn, Đóng góp)

---

## 🔐 Bảo mật

TeleCloud mã hóa dữ liệu nhạy cảm trong DB và chạy với các tiêu chuẩn hardening cơ bản, nhưng vẫn cần thao tác tay từ người vận hành.

**Bắt buộc:**

*   **`TELECLOUD_MASTER_KEY`** — 32 byte ngẫu nhiên (`openssl rand -hex 32`). Dùng để mã hóa Telegram session blob và các setting nhạy cảm (`api_id`, `api_hash`, `log_group_id`, `bot_tokens`) bằng AES-256-GCM. **Sao lưu key này tách biệt với DB**. Mất key = không thể giải mã, phải re-auth Telegram và nhập lại bot token. Script `auto-setup.sh` tự sinh và ghi vào `.env`; bạn cần copy ra password manager.

**Khuyến nghị:**

*   **`TELECLOUD_SETUP_TOKEN`** — token một lần để chặn scanner cướp `/setup` trước khi bạn kịp tạo admin. Sinh bằng `openssl rand -hex 16`. Khi đã set, request tới `/setup` phải kèm `X-Setup-Token` hoặc `?token=<value>`.
*   **`LISTEN_ADDR`** — mặc định `127.0.0.1` khi chưa setup admin (chống lộ wizard ra mạng công cộng), `0.0.0.0` sau khi setup. Đặt sau Cloudflare Tunnel / Nginx / Tailscale thay vì mở port trực tiếp.
*   **Docker Compose** mặc định binding `127.0.0.1:8091`. Đổi sang `0.0.0.0` chỉ nếu bạn cố tình muốn lộ ra ngoài.

**Đã làm sẵn:**

*   Telegram session + các setting nhạy cảm mã hóa AES-256-GCM khi lưu DB; migrate tự động trên boot (SQLite tự tạo `.pre-enc-<ts>.bak`; với MySQL/Postgres yêu cầu `TELECLOUD_I_HAVE_BACKED_UP=1` để chạy).
*   `auto-setup.sh` tạo systemd unit với user riêng (`User=telecloud` hoặc `DynamicUser`), `NoNewPrivileges`, `ProtectSystem=strict`, `ProtectHome`, `PrivateTmp` và verify SHA-256 binary từ `checksums.txt` cùng release.
*   Token chia sẻ trực tiếp dùng HMAC key derived từ master key (không phụ thuộc admin password — đổi password không invalidate link đã phát hành).
*   Share password protection dùng signed session token, **không** lưu bcrypt hash trong cookie.
*   WebDAV có rate limit 5/15 phút/IP cho lần auth sai; cache verify ngắn (2 phút) và key dùng SHA-256 của mật khẩu, không phải plaintext.
*   Session HTTP có `expires_at` (mặc định 30 ngày). Đổi password sẽ invalidate mọi session khác của user. Cleanup tự động 6 tiếng.
*   Audit log (`audit_log` table) ghi login/logout, đổi password, admin reset, setup, thay đổi setting. Lưu 90 ngày.
*   Tải URL từ xa (remote upload, yt-dlp...) dùng `SafeHTTPClient` có pinned dial — chặn DNS rebinding tới private/loopback IP.
*   CSP mặc định **không** cho phép Cloudflare Web Analytics. Bật setting `analytics_enabled=true` nếu muốn opt-in. CSP vẫn cần `'unsafe-inline'` + `'unsafe-eval'` vì frontend dùng Alpine.js.

**Hạn chế đã biết:**

*   Telegram cloud chat **không** end-to-end encrypted. Đừng lưu giấy tờ tùy thân, ảnh riêng tư, secret keys hoặc tài liệu nhạy cảm vào TeleCloud — encrypt phía client (`rclone crypt`, Cryptomator) trước khi upload nếu vẫn muốn dùng làm storage.
*   Telegram có quyền ban tài khoản dùng userbot cho mục đích lưu trữ; không có recovery nếu bị ban. Dùng số điện thoại + tài khoản phụ.

---

## ⚠️ Điều khoản sử dụng & Miễn trừ trách nhiệm

Dự án **TeleCloud** được phát triển nhằm mục đích lưu trữ và quản lý tệp tin cá nhân hợp pháp. Chúng tôi không chịu trách nhiệm đối với bất kỳ nội dung nào được người dùng tải lên hoặc các vi phạm điều khoản sử dụng của Telegram. Người dùng **hoàn toàn tự chịu trách nhiệm** cho hành vi sử dụng của mình.

Dự án được cung cấp **"nguyên trạng" (as-is)**, không có bất kỳ đảm bảo nào về tính ổn định hay bảo mật.

---

## 🙏 Đóng góp

Dự án sử dụng các thư viện tuyệt vời: 
* [gotd/td](https://github.com/gotd/td): Telegram client, in Go. (MTProto API)
* [Gin](https://github.com/gin-gonic/gin): High-performance HTTP web framework.
* [AlpineJS](https://github.com/alpinejs/alpine): A rugged, minimal framework for JS.
* [TailwindCSS](https://github.com/tailwindlabs/tailwindcss): A utility-first CSS framework.
* [plyr](https://github.com/sampotts/plyr): A simple HTML5 media player.
* [Artplayer.js](https://github.com/zhw2590582/ArtPlayer): Modern and full-featured HTML5 video player.
* [Prism.js](https://github.com/PrismJS/prism): Lightweight, extensible syntax highlighter.
* [FontAwesome](https://fontawesome.com): Bộ biểu tượng phổ biến nhất thế giới.
* [yt-dlp](https://github.com/yt-dlp/yt-dlp): Audio/video downloader.
* [aria2](https://github.com/aria2/aria2): Multi-protocol download utility.
* [Google Fonts (Nunito)](https://fonts.google.com/specimen/Nunito): Modern sans-serif typeface.

Xin cảm ơn các đội ngũ phát triển và các **nhà đóng góp (contributors)** đã cung cấp những công cụ và nỗ lực hữu ích cho cộng đồng.

<a href="https://github.com/dabeecao/telecloud-go/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=dabeecao/telecloud-go" />
</a>

**Một phần mã nguồn của dự án và readme này được tham khảo và chỉnh sửa bởi Gemini AI**

---

## 📜 Giấy phép

Dự án này được phát hành dưới giấy phép [GNU Affero General Public License v3.0 (AGPL-3.0)](https://www.gnu.org/licenses/agpl-3.0.html).
