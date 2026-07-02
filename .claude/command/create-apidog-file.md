# Create Apidog File

Router dosyasından otomatik Apidog JSON dosyası oluştur.

## Kurallar

Türkçe dökümantasyon olmalı.

## Kullanım

```
/create-apidog-file <router_file_path>
```

**ÖNEMLİ:** Parametre olarak **MUTLAKA router dosyasının tam (absolute) path'ini** verin!

✅ **DOĞRU:**

```
/create-apidog-file /Users/fatihdemir/Desktop/soberx_backend/internal/api/apiroots/auth.go
```

❌ **YANLIŞ:**

```
/create-apidog-file /Users/fatihdemir/Desktop/soberx_backend/internal/models/User.go
```

(Model dosyası değil, router dosyası verin!)

## Örnek

```
/create-apidog-file /Users/fatihdemir/Desktop/soberx_backend/internal/api/apiroots/auth.go
```

## Görev

Verilen router dosyasındaki **TÜM endpoint'ler** için Apidog formatında **TEK BİR** JSON dosyası oluştur.

**ÖNEMLİ**: Router dosyasında kaç entity olursa olsun (örn: User, Plan, Entitlements, Purchase), hepsi için tüm endpoint'leri (Liste, Kaydı Getir, Ekle, Güncelle, Sil) tek bir Apidog JSON dosyasına ekle.

### Adımlar

1. **Router Dosyasını Analiz Et**
    - Router dosyasını oku
    - Tüm endpoint tanımlarını bul (GET, POST, PUT, DELETE)
    - Her endpoint için:
        - HTTP method
        - Path
        - Controller fonksiyon adı

2. **Controller Dosyalarını Bul ve Analiz Et**
    - Her endpoint'in controller fonksiyonunu bul
    - Controller dosyasını oku
    - Kod analizi yap:
        - Request body struct'ını bul (c.ShouldBindJSON ile bind edilen)
        - Service fonksiyonunu tespit et (örn: `service.Register(request)`)
        - Service fonksiyonunun return type'ını belirle (BaseServiceResponse)
        - Response data field'ının model tipini bul (örn: models.User)
        - Swagger dökümanlarını oku (@Summary, @Description, @Tags, @Param, @Success)
          
3. **Model Dosyalarını Detaylı Analiz Et**
    - Controller'dan belirlenen model dosyasını oku (örn: `/Users/fatihdemir/Desktop/soberx_backend/internal/models/User.go`)
    - Her field için şu bilgileri çıkar:
        - Field adı ve JSON tag'i (`json:"field_name"`)
        - Field tipi (string, int, float64, time.Time, uuid.UUID, embedded struct, vb.)
        - **Description tag'i** (`description:"Alan açıklaması"`)
        - Required bilgisi (json tag'inde `omitempty` yoksa required)
    - Description bilgisi için öncelik sırası:
        1. Struct tag'indeki `description` değeri (örn: `description:"El Hijyeni Numunesi Adı"`)
        2. Field üzerindeki Go comment satırı (örn: `// El Hijyeni Numunesi Adı`)
        3. Eğer ikisi de yoksa, field adından türet
    - Embedded struct'lar için (Lookup ve Picklist field'lar):
        - Lookup field'lar: `BaseRecordFields` → `{id: uuid, name: string}` yapısı
        - Picklist field'lar: `{value: number, text: string}` yapısı ve enum değerleri
    - Field sırasını koru (struct'taki sıraya göre)

4. **JSON Schema Oluştur**
    - Go struct'larını JSON Schema'ya dönüştür
    - **Her field için description ekle** (Adım 3'te çıkarılan description bilgisini kullan)
    - Field tipleri:
        - `string` → "type": "string"
        - `int`, `int64`, `float64` → "type": "number"
        - `bool` → "type": "boolean"
        - `time.Time` → "type": "string", "format": "date-time"
        - `uuid.UUID` → "type": "string", "format": "uuid"
        - Embedded struct → "type": "object"
    - Required field'ları belirle (json:"field" tag'inde `omitempty` yoksa required)
    - x-apidog-orders array'ini field sırası ile oluştur (struct'taki sırayla aynı)

   **Request Schema Örneği:**
   ```json
   {
     "type": "object",
     "properties": {
       "first_name": {
         "type": "string",
         "description": "Kullanıcının adı",
         "example": "Fatih"
       },
       "last_name": {
         "type": "string",
         "description": "Kullanıcının soy adı",
         "example": "Demir"
       },
       "email": {
         "type": "string",
         "description": "Kullanıcının email adresi",
         "example": "fath.demmr@gmail.com"
         },
       "password": {
         "type": "string",
         "description": "Kullanıcının şifresi",
         "example": "StrongPassword123!"
        }  
      }
   }
   ```

   **Response Schema Örneği:**
   ```json
   {
     "type": "object",
     "properties": {
       "success": {
         "type": "boolean",
         "example": true
       },
       "message": {
         "type": "string",
         "example": "User registered successfully"
       },
       "code": {
         "type": "string",
         "example": "200"
       },
       "data": {
         "type": "object",
         "properties": {
           "id": {
             "type": "string",
             "example": "3f5c9f4a-7a22-4b4d-9d8e-8c7f2f1a1111"
           },
           "apple_id": {
             "type": "string",
             "example": ""
           },
           "first_name": {
             "type": "string",
             "example": "Fatih"
           },
           "last_name": {
             "type": "string",
             "example": "Demir"
           },
           "full_name": {
             "type": "string",
             "example": "Fatih Demir"
           },
           "email": {
             "type": "string",
             "format": "email",
             "example": "fatih@example.com"
           },
           "is_email_verified": {
             "type": "boolean",
             "example": true
           },
           "password": {
             "type": "string",
             "example": "$2a$10$hashedpasswordvalue"
           },
           "login_type": {
             "type": "integer",
             "example": 1
           },
           "login_type_name": {
             "type": "string",
             "example": "EMAIL"
           },
           "phone_number": {
             "type": "string",
             "example": "+905551112233"
           },
           "last_login_time": {
             "type": "string",
             "format": "date-time",
             "example": "2026-05-12T14:30:00Z"
           },
           "client_ip": {
             "type": "string",
             "example": "192.168.1.1"
           },
           "created_at": {
             "type": "string",
             "format": "date-time",
             "example": "2026-05-12T14:00:00Z"
           },
           "created_by": {
             "type": "integer",
             "example": 1
           },
           "created_by_name": {
             "type": "string",
             "example": "System Admin"
           },
           "created_ip_address": {
             "type": "string",
             "example": "192.168.1.10"
           },
           "updated_at": {
             "type": "string",
             "format": "date-time",
             "example": "2026-05-12T14:30:00Z"
           },
           "updated_by": {
             "type": "integer",
             "example": 1
           },
           "updated_by_name": {
             "type": "string",
             "example": "System Admin"
           },
           "updated_ip_address": {
             "type": "string",
             "example": "192.168.1.10"
           }
         }
       }
     }
   }
   ```

5. **Markdown Döküman Oluştur**
   Her endpoint için detaylı markdown döküman:
   ```markdown
   # {Endpoint Türkçe Adı}

   ## Genel Bakış
   {Swagger @Description}

   ## HTTP Method & Path
   `{METHOD} {PATH}`

   ## Request Body

   ### Gerekli Alanlar
   - **{field_name}** ({type}): {description}

   ### Opsiyonel Alanlar
   - **{field_name}** ({type}): {description}

   ## Response

   ### Success (200/201)
   ```json
   {
     "success": true,
     "message": "",
     "data": { ... }
   }
   ```

   ### Error (400/500)
   ```json
   {
     "success": false,
     "message": "error message",
     "code": "..."
   }
   ```

   ## Örnek Request
   ```json
   { ... }
   ```

   ## Notlar
    - Authentication gereklidir
   ```

6. **Apidog JSON Dosyası Oluştur**

   **ÇOK ÖNEMLİ - DOSYA FORMATI UYARISI:**
    - Dosya formatı **KESINLIKLE** Apidog formatında olmalıdır
    - **ASLA** OpenAPI formatı ("openapi": "3.0.1") kullanma!
    - İlk satır **MUTLAKA** `"apidogProject": "1.0.0"` olmalıdır
    - OpenAPI formatı ile Apidog formatı farklıdır, karıştırma!

   Referans format:

   Doğru Apidog Yapısı:
   ```json
   {
     "apidogProject": "1.0.0",
     "$schema": {
       "app": "apidog",
       "type": "project",
       "version": "1.2.0"
     },
     "info": {
       "name": "{Entity Name} API",
       "description": "Auto-generated from router file",
       "mockRule": {
         "rules": [],
         "enableSystemRule": true
       }
     },
     "apiCollection": [
       {
         "name": "Root",
         "id": {random_id},
         "items": [
           {
             "name": "Hygiene Api",
             "items": [
               {
                 "name": "{Entity Name}",
                 "items": [
                   {
                     "name": "Liste",
                     "api": {
                       "id": "{random_id}",
                       "method": "get",
                       "path": "/api/...",
                       "requestBody": { ... },
                       "responses": [ ... ],
                       "description": "{markdown_content}"
                     }
                   },
                   {
                     "name": "Ekle",
                     "api": { ... }
                   },
                   {
                     "name": "Kaydı Getir",
                     "api": { ... }
                   },
                   {
                     "name": "Güncelle",
                     "api": { ... }
                   },
                   {
                     "name": "Sil",
                     "api": { ... }
                   }
                 ]
               }
             ]
           }
         ]
       }
     ]
   }
   ```

7. **Dosyayı Kaydet**
    - Hedef dizin: `/Users/fatihdemir/Desktop/soberx_backend/apidog/<entity_name>`
    - Dosya adı: `{entity_name}_api.json` (örn: `samples_api.json`, `water_hygiene_sample_api.json`)

### Endpoint Türkçe İsimlendirme

HTTP Method'a göre Türkçe isim:

- GET (liste) → "Liste"
- GET (by id) → "Kaydı Getir"
- POST → "Ekle"
- PUT → "Güncelle"
- DELETE → "Sil"

### Önemli Notlar

- Her endpoint için benzersiz ID üret (random 8 haneli sayı)
- JSON Schema'da field sıralamasını koru (x-apidog-orders)
- **Her field için description ekle** (model struct'ından description tag'ini al)
- Lookup field'lar için nested object yapısı kullan
    - `location.id` için description: "Lokasyon ID'si"
    - `location.name` için description: "Lokasyon adı"
    - `location` object için description: Model struct'taki description tag'inden al
- Picklist field'lar için value + text yapısı kullan
    - `value` için description: Enum değerlerini ve anlamlarını yaz (örn: "100000000: -, 100000001: Pass, 100000002:
      Fail")
    - `text` için description: "Sonuç metni" gibi açıklayıcı bir metin
    - Picklist object için description: Model struct'taki description tag'inden al
- Required field'ları doğru belirle
- Response schema için BaseServiceResponse yapısını kullan
    - `success` için description: "İşlem başarı durumu"
    - `message` için description: "İşlem mesajı"
    - `data` için: Model struct'ındaki tüm field'ların description'larını ekle
- **ÇOK ÖNEMLİ: Auth bloğu EKLEME** - Her endpoint'e ayrı ayrı `auth` bloğu ekleme! Authentication ayarları inherit
  olarak yapılacak, bu yüzden endpoint tanımlarında `auth` property'si olmamalı.
- **ÇOK ÖNEMLİ: ASLA $ref kullanma** - Tüm response ve request schema'larında inline (yerinde) schema tanımla. `$ref`
  veya `#/definitions/` kullanma!
- **ÇOK ÖNEMLİ: Liste endpoint'lerinde array items detaylı olmalı**:
    - Liste endpoint'lerinde (GET liste) response içindeki array (örn: `roles`, `permissions`, vb.) için `items`
      kısmında TÜM model field'larını inline olarak yaz
    - Örnek:
      `"roles": { "type": "array", "items": { "type": "object", "properties": { "id": {...}, "name": {...}, ... TÜM FIELD'LAR } } }`
    - Array items'ı boş bırakma veya sadece `"$ref"` ile referans verme!
- **ÇOK ÖNEMLİ: Response data her zaman detaylı olmalı**:
    - Detay endpoint'lerinde (GET by id, GET by code, POST create, PUT update) response'daki `data` object'inin
      `properties`'ini tam ve detaylı yaz
    - `data` için sadece `"type": "object"` deyip geçme! Tüm field'ları `properties` içinde inline olarak tanımla
    - Her field için description, type, format (uuid, date-time vb.) belirt
- Tüm endpoint'leri "developing" status ile oluştur
- ordering değerleri: Liste=10, Kaydı Getir=15, Ekle=20, Güncelle=30, Sil=40

### Örnek Output Structure

```
/Users/fatihdemir/Desktop/soberx_backend/apidog/
├── samples_api.json
│   ├── Hand Hygiene Sample (5 endpoints)
│   ├── Ice Hygiene Sample (5 endpoints)
│   ├── Surface Hygiene Sample (5 endpoints)
│   ├── Water Hygiene Sample (5 endpoints)
│   └── Air Hygiene Sample (5 endpoints)
```

## Context

{{CONTEXT}}