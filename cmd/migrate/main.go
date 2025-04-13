package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// .env dosyasını yükle
	err := godotenv.Load()
	if err != nil {
		log.Printf("Uyarı: .env dosyası yüklenemedi: %v", err)
	}

	// Environment değişkenlerini oku
	dbURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	backupPath := os.Getenv("BACKUPS_PATH")

	// Gerekli environment değişkenlerini kontrol et
	if dbURL == "" {
		log.Fatal("HATA: DATABASE_URL environment değişkeni ayarlanmamış.")
	}
	if migrationsPath == "" {
		log.Fatal("HATA: MIGRATIONS_PATH environment değişkeni ayarlanmamış.")
	}
	if backupPath == "" {
		log.Fatal("HATA: BACKUP_PATH environment değişkeni ayarlanmamış.")
	}

	// Migration kaynağını ve veritabanı bağlantısını hazırla (backup/push hariç komutlar için)
	var m *migrate.Migrate
	sourceURL := "file://" + migrationsPath

	// Komut satırı argümanlarını işle
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("HATA: Komut belirtilmedi.")
	}
	command := args[0]

	if command != "backup-pull" && command != "backup-push" {
		var migrateErr error
		m, migrateErr = migrate.New(sourceURL, dbURL)
		if migrateErr != nil {
			log.Fatalf("HATA: Migrate başlatılamadı: %v", migrateErr)
		}
		defer func() {
			srcErr, dbErr := m.Close()
			if srcErr != nil {
				log.Printf("Uyarı: Kaynak kapatılırken hata: %v", srcErr)
			}
			if dbErr != nil {
				log.Printf("Uyarı: Veritabanı bağlantısı kapatılırken hata: %v", dbErr)
			}
		}()
	}

	var version int
	switch command {
	case "backup-pull":
		fmt.Println("Bilgi: Mevcut Veritabanı yedekleniyor...")
		err = backupDatabase(dbURL, backupPath)
		if err == nil {
			fmt.Println("Başarılı: Veritabanı başarıyla yedeklendi.")
		}
	case "backup-push":
		fmt.Println("Bilgi: En güncel yedek dosyası veritabanına yükleniyor...")
		fmt.Println("UYARI: Bu işlem mevcut veritabanını en son yedekle değiştirecektir.")
		fmt.Print("Devam etmek istiyor musunuz? (y/N): ")
		var confirmation string
		fmt.Scanln(&confirmation)
		if strings.ToLower(confirmation) != "y" {
			fmt.Println("İşlem iptal edildi.")
			return
		}

		err = restoreDatabase(dbURL, backupPath)
		if err == nil {
			fmt.Println("Başarılı: Veritabanı başarıyla geri yüklendi.")
		}
	case "up":
		fmt.Println("Bilgi: Bir adım ileri migration uygulanıyor...")
		err = m.Steps(1)
		if err == nil {
			fmt.Println("Başarılı: Bir adım ileri migration uygulandı.")
		}
	case "down":
		fmt.Println("Bilgi: Bir adım geri migration alınıyor...")
		err = m.Steps(-1)
		if err == nil {
			fmt.Println("Başarılı: Bir adım geri migration alındı.")
		}
	case "up_all":
		fmt.Println("Bilgi: Tüm bekleyen migration'lar uygulanıyor...")
		err = m.Up()
		if err == nil {
			fmt.Println("Başarılı: Tüm bekleyen migration'lar uygulandı.")
		}
	case "down_all":
		fmt.Println("Bilgi: Tüm migration'lar geri alınıyor...")
		err = m.Down()
		if err == nil {
			fmt.Println("Başarılı: Tüm migration'lar geri alındı.")
		}
	case "force":
		if len(args) < 2 {
			log.Fatal("HATA: 'force' komutu için versiyon numarası gerekli.")
		}
		version, err = strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("HATA: Geçersiz versiyon numarası: %s", args[1])
		}
		fmt.Printf("Bilgi: Migration versiyonu %d olarak zorlanıyor...\n", version)
		err = m.Force(version)
		if err == nil {
			fmt.Printf("Başarılı: Migration versiyonu %d olarak zorlandı.\n", version)
		}
	case "goto":
		if len(args) < 2 {
			log.Fatal("HATA: 'goto' komutu için versiyon numarası gerekli.")
		}
		versionUint64, errConv := strconv.ParseUint(args[1], 10, 64)
		if errConv != nil {
			log.Fatalf("HATA: Geçersiz versiyon numarası: %s", args[1])
		}
		version = int(versionUint64) // Sadece yazdırma için
		fmt.Printf("Bilgi: Migration versiyon %d hedefine gidiliyor...\n", version)
		err = m.Migrate(uint(versionUint64))
		if err == nil {
			fmt.Printf("Başarılı: Migration versiyon %d hedefine ulaşıldı.\n", version)
		}
	case "version":
		v, dirty, errVersion := m.Version()
		if errVersion != nil {
			if errors.Is(errVersion, migrate.ErrNilVersion) {
				fmt.Println("Bilgi: Henüz uygulanmış migration yok.")
				return // Hata yok, çık
			}
			log.Fatalf("HATA: Versiyon alınamadı: %v", errVersion)
		}
		fmt.Printf("Mevcut Versiyon: %d\nDirty Durumu: %t\n", v, dirty)
		return // Hata yok, çık
	default:
		log.Fatalf("HATA: Geçersiz komut: %s", command)
	}

	// Genel hata kontrolü (backup/restore dahil)
	if err != nil {
		// migrate paketine özgü hataları kontrol et
		if m != nil { // Sadece migrate komutları için bu kontrolü yap
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("Bilgi: Değişiklik olmadı (No change).")
				return // Hata yok, çıkış yap
			} else if errors.Is(err, migrate.ErrNilVersion) && (command == "down" || command == "down_all") {
				fmt.Println("Bilgi: Geri alınacak migration bulunamadı (ErrNilVersion).")
				return // Hata yok, çıkış yap
			}
		}
		// Diğer tüm hatalar (backup/restore hataları dahil)
		log.Fatalf("HATA: Komut '%s' çalıştırılırken hata oluştu: %v", command, err)
	}
}

// backupDatabase fonksiyonu pg_dump kullanarak veritabanını yedekler.
func backupDatabase(dbURL, backupDir string) error {
	// Backup dizininin var olduğundan emin ol, yoksa oluştur.
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("yedekleme dizini oluşturulamadı '%s': %w", backupDir, err)
	}

	// Zaman damgalı dosya adı oluştur
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("backup_%s.sql", timestamp)
	backupFilePath := filepath.Join(backupDir, backupFileName)

	// pg_dump komutunu oluştur
	// Önemli: dbURL'nin pg_dump tarafından doğru yorumlanması için tırnak içine alınması gerekebilir.
	// exec.Command bunu genellikle halleder, ancak karmaşık URL'lerde sorun olabilir.
	cmd := exec.Command("pg_dump", dbURL, "-f", backupFilePath, "--clean") // --clean eklemek restore öncesi drop komutları ekler

	// Komutun çıktısını ve hatalarını yakala
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Çalıştırılıyor: %s\n", cmd.String()) // Çalıştırılan komutu logla

	// Komutu çalıştır
	err := cmd.Run()
	if err != nil {
		// Başarısız olursa oluşturulan boş dosyayı silmeye çalış
		_ = os.Remove(backupFilePath)
		return fmt.Errorf("pg_dump komutu başarısız oldu: %w", err)
	}

	log.Printf("Yedekleme başarıyla oluşturuldu: %s\n", backupFilePath)
	return nil
}

// restoreDatabase fonksiyonu psql kullanarak en son yedeği geri yükler.
func restoreDatabase(dbURL, backupDir string) error {
	// En son backup dosyasını bul
	latestBackupFile, err := getLatestBackupFile(backupDir, ".sql")
	if err != nil {
		return fmt.Errorf("en son yedek dosyası bulunamadı: %w", err)
	}
	if latestBackupFile == "" {
		return errors.New("yedekleme dizininde '.sql' uzantılı yedek dosyası bulunamadı")
	}

	backupFilePath := filepath.Join(backupDir, latestBackupFile)
	log.Printf("Geri yüklenecek yedek dosyası: %s\n", backupFilePath)

	// psql komutunu oluştur
	// -1 veya --single-transaction: Tüm komutları tek bir transaction içinde çalıştırır.
	// -v ON_ERROR_STOP=1: Herhangi bir SQL hatasında işlemi durdurur.
	// -f: Belirtilen dosyadan komutları okur.
	// dbURL: Bağlanılacak veritabanı URL'si.
	cmd := exec.Command("psql", dbURL, "-v", "ON_ERROR_STOP=1", "-f", backupFilePath)

	// Komutun çıktısını ve hatalarını yakala
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("Çalıştırılıyor: %s\n", cmd.String()) // Çalıştırılan komutu logla

	// Komutu çalıştır
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("psql komutu başarısız oldu: %w", err)
	}

	log.Printf("Veritabanı başarıyla geri yüklendi: %s\n", backupFilePath)
	return nil
}

// getLatestBackupFile belirtilen dizindeki en son değiştirilen dosyayı bulur.
func getLatestBackupFile(dir, extension string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("dizin okunamadı '%s': %w", dir, err)
	}

	var backupFiles []os.FileInfo
	for _, file := range files {
		// Sadece belirtilen uzantıya sahip dosyaları ve dizin olmayanları al
		if !file.IsDir() && strings.HasSuffix(file.Name(), extension) {
			info, err := file.Info()
			if err != nil {
				log.Printf("Uyarı: Dosya bilgisi alınamadı '%s': %v", file.Name(), err)
				continue // Bu dosyayı atla
			}
			backupFiles = append(backupFiles, info)
		}
	}

	if len(backupFiles) == 0 {
		return "", nil // Yedek dosyası bulunamadı
	}

	// Dosyaları değiştirilme zamanına göre sırala (en yeni en sonda)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].ModTime().Before(backupFiles[j].ModTime())
	})

	// En son dosyayı döndür
	return backupFiles[len(backupFiles)-1].Name(), nil
}
