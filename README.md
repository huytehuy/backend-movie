# Movie Watch Party - Go Backend

Backend API cho ứng dụng xem phim với tính năng watch party (xem cùng nhau) sử dụng Golang và WebSocket.

## Tính năng

### 🎬 Video Streaming
- HTTP range requests hỗ trợ tua video (seeking)
- Streaming video với hiệu suất cao
- Upload video qua API

### 🎉 Watch Party
- Tạo phòng xem phim cùng nhau
- Đồng bộ video real-time (play/pause/seek) qua WebSocket
- Danh sách người dùng trong phòng
- Chat trong phòng (tùy chọn)

### 🛠️ Utilities
- Tự động tạo thumbnail từ video
- Transcode video sang định dạng web-friendly
- Lấy thông tin video (duration, metadata)

## Cài đặt

### Yêu cầu
- Go 1.21+
- FFmpeg (cho thumbnail và transcoding)

### Cài đặt dependencies
```bash
go mod download
```

### Tạo thư mục cần thiết
```bash
mkdir videos thumbnails
```

## Chạy server

```bash
go run .
```

Server sẽ chạy trên `http://localhost:8080`

## API Endpoints

### Movies
- `GET /api/movies` - Lấy danh sách phim
- `GET /api/movies/{id}` - Lấy thông tin phim
- `POST /api/upload` - Upload video mới

### Video Streaming
- `GET /api/videos/{filename}` - Stream video (hỗ trợ range requests)
- `GET /api/thumbnails/{filename}` - Lấy thumbnail

### Watch Party
- `POST /api/rooms` - Tạo phòng mới
  ```json
  {
    "movieId": "1",
    "roomName": "My Party Room",
    "username": "John"
  }
  ```
- `GET /api/rooms/{id}` - Lấy thông tin phòng
- `WS /api/rooms/{id}/ws?username={name}` - WebSocket kết nối

### Health
- `GET /api/health` - Health check

## WebSocket Messages

### Client -> Server

**Play**
```json
{
  "type": "play",
  "data": {
    "currentTime": 123.45
  }
}
```

**Pause**
```json
{
  "type": "pause",
  "data": {
    "currentTime": 123.45
  }
}
```

**Seek**
```json
{
  "type": "seek",
  "data": {
    "time": 200.0
  }
}
```

**Chat**
```json
{
  "type": "chat",
  "data": {
    "message": "Hello everyone!"
  }
}
```

### Server -> Client

**Sync (Video State)**
```json
{
  "type": "sync",
  "roomId": "abc123",
  "data": {
    "isPlaying": true,
    "currentTime": 123.45,
    "lastUpdateBy": "John",
    "updatedAt": "2026-02-09T14:00:00Z"
  },
  "timestamp": "2026-02-09T14:00:00Z"
}
```

**User List**
```json
{
  "type": "userList",
  "roomId": "abc123",
  "data": [
    {"id": "user1", "username": "John"},
    {"id": "user2", "username": "Jane"}
  ],
  "timestamp": "2026-02-09T14:00:00Z"
}
```

**Play/Pause/Seek Events**
```json
{
  "type": "play",
  "roomId": "abc123",
  "userId": "user1",
  "username": "John",
  "data": {
    "currentTime": 123.45
  },
  "timestamp": "2026-02-09T14:00:00Z"
}
```

## Cấu trúc Project

```
movieapp/
├── main.go          # Entry point, router setup
├── server.go        # HTTP handlers (movies, video streaming)
├── party.go         # WebSocket server, room management
├── models.go        # Data structures
├── transcode.go     # Video processing utilities
├── go.mod           # Go modules
├── videos/          # Video files
└── thumbnails/      # Video thumbnails
```

## Tích hợp với React Frontend

Backend này được thiết kế để hoạt động với React frontend. Để tích hợp:

1. **Development**: CORS đã được cấu hình cho `localhost:5173` (Vite) và `localhost:3000` (CRA)

2. **Production**: Build React app và uncomment dòng trong `main.go`:
   ```go
   router.PathPrefix("/").Handler(http.FileServer(http.Dir("./movie/dist")))
   ```

3. **Environment Variables**: 
   - `PORT`: Server port (default: 8080)

## Upload Video

Sử dụng `curl` hoặc Postman:

```bash
curl -X POST http://localhost:8080/api/upload \
  -F "video=@path/to/video.mp4"
```

Video sẽ được lưu vào thư mục `videos/` và thumbnail tự động được tạo.

## License

MIT
