# Product Requirement Document (PRD) - FinFlow Wealthboard

* **Nama Produk**: FinFlow Wealthboard
* **Status**: Draf Kebutuhan (Diperbarui dengan Rekomendasi Sistem Pencatatan Ideal)
* **Target Pengguna**: Investor retail yang membutuhkan pencatatan keuangan mandiri terintegrasi
* **Arsitektur**: Aplikasi Web Lokal (Offline-first, Golang + Vanilla Web)

---

## 1. Latar Belakang & Masalah (Background & Problem Statement)

Saat ini, pengelolaan keuangan dan investasi pribadi Anda tersebar di berbagai platform yang terpisah:
1.  **Notion** untuk mencatat transaksi harian, mengevaluasi anggaran bulanan, dan melacak portofolio aset.
2.  **Aplikasi Sekuritas/Aplikasi Agen Penjual Reksa Dana (Stockbit, Bibit, BNI)** untuk eksekusi transaksi saham, reksa dana, dan emas.
3.  **Gemini AI** untuk melakukan konsultasi, analisis taktis rebalancing, dan evaluasi pengeluaran mingguan/bulanan.

**Masalah Utama & Friction di Notion:**
*   **Saham Pecahan (Fractional Shares)**: Karena rumus `Unit = Nominal / Harga Beli` memasukkan fee transaksi ke dalam nominal beli, unit saham tercatat sebagai pecahan (misal: 800.801269 saham). Di pasar saham riil (IDX), saham hanya diperdagangkan dalam satuan lot utuh (1 lot = 100 lembar).
*   **Pembaruan Manual yang Melelahkan**: Pembaruan harga pasar saat ini (*Current Market Price*) untuk saham dan emas harus dihitung dan diinput manual di Notion untuk mengetahui kinerja portofolio berjalan (*floating profit/loss*).
*   **Kalkulasi Penjualan Aset**: Sistem rollup Notion yang terbatas pada penjumlahan (SUM) memaksa pengguna menginput nominal negatif untuk transaksi penjualan (*Sell*), yang tidak intuitif.
*   **Koneksi Lambat & Konteks Relasi**: Notion membutuhkan koneksi internet dan pemuatan relasi bulanan manual yang memperlambat proses pencatatan harian.

---

## 2. Tujuan Produk (Product Objectives)

*   **Penyatuan Data**: Mengintegrasikan pencatatan pemasukan, pengeluaran, anggaran bulanan, pelacakan investasi, dan evaluasi berkala dalam satu aplikasi terpadu.
*   **Kepatuhan Privasi (100% Lokal)**: Menjamin data keuangan tetap berada di mesin lokal dengan membaca/menulis langsung dari file CSV/JSON lokal.
*   **Otomatisasi Analisis & Integrasi Riil**: Mengotomatiskan penarikan harga saham IDX (via Yahoo Finance API) secara real-time dan kalkulasi portfolio tanpa unit pecahan.
*   **Pengalaman Pengguna Premium (Keyboard-First)**: Menghadirkan antarmuka visual modern bertema gelap (*Glassmorphic Dark Mode*) yang cepat (<200ms) dan dapat dioperasikan penuh menggunakan keyboard untuk mencatat transaksi secara instan.

---

## 3. Fitur Utama & Kebutuhan Fungsional (Core Features & Requirements)

### A. Dashboard Utama (Wealth Dashboard)
*   **Kartu Aset Utama**: Menampilkan total kekayaan bersih (*Net Worth*), total kas/bank, investasi reksa dana/emas, dan total nilai saham terkini.
*   **Grafik Distribusi Aset**: Pie chart interaktif Chart.js untuk pembagian kelas aset (Kas, RDPU, RDPT, Saham, Emas).
*   **Rasio Kesehatan Finansial (Financial Health Ratios)**:
    *   *Savings Rate*: Persentase uang yang diinvestasikan/ditabung dari total pendapatan bulan berjalan.
    *   *Emergency Fund Run-Rate*: Estimasi berapa bulan dana darurat Anda saat ini dapat menutupi biaya hidup (dihitung otomatis dari rata-rata pengeluaran bulanan).
*   **Satpam Virtual (Visual Alerts)**: Widget di layar yang memberikan peringatan visual jika harga saham saat ini menyentuh batas harga beli/jual yang ditentukan pengguna (misal: BBRI menyentuh batas Rp2.800).
*   **Peringatan Anggaran**: Indikator visual jika pengeluaran kategori anggaran tertentu telah melebihi 80% limit atau mengalami *over-budget*.

### B. Pencatat Arus Kas (Cashflow Tracker & Quick Input)
*   **Input Transaksi Cepat (Keyboard-First)**:
    *   Form input nominal (otomatis format Rupiah saat mengetik), tanggal (default hari ini), toko/sumber, kategori tag, dan metode pembayaran.
    *   Mendukung penuh navigasi keyboard (tombol `Tab` untuk berpindah field, dan `Enter` untuk menyimpan) agar transaksi tercatat dalam <5 detik.
*   **Pencatatan Dividen & Kupon**: Form khusus untuk mencatat dividen saham atau kupon obligasi. Input ini akan otomatis:
    1.  Menambah kas di `income_v2.csv` (Kategori: Dividends/Interest).
    2.  Tercatat di jurnal investasi untuk melacak *Realized Return* portofolio terkait.
*   **Sinkronisasi File**: Transaksi langsung disimpan sebagai baris baru di file `income_v2.csv` atau `expenses_v2.csv`. Sistem lokal otomatis mem-parsing tanggal transaksi untuk pengelompokan bulanan (tanpa perlu relasi bulan manual).
*   **Tabel Transaksi Terbaru**: Tabel interaktif dengan fitur pencarian dan filter berdasarkan tag kategori atau tanggal.

### C. Pengendali Anggaran (Budget Planner)
*   **Batas Anggaran**: Menentukan batas anggaran bulanan per kategori (Groceries, Bensin, Jajan, Listrik, dll).
*   **Progress Bar Realisasi**: Visualisasi dinamis realisasi anggaran dengan skema warna (Hijau: aman, Jingga: mendekati limit, Merah: over-budget).
*   **Kalkulator Sisa Anggaran**: Menghitung sisa anggaran riil per kategori secara otomatis berdasarkan pengeluaran terbaru.

### D. Manajemen Portofolio Investasi (Portfolio Tracker)
*   **Integrasi Multi-Aset**: Tab khusus untuk RDPU, RDPT, Emas, dan Saham.
*   **Logika Jurnal Transaksi Portofolio Ideal**:
    *   Input transaksi wajib mencakup: **Quantity (Jumlah Lembar/Unit)** dan **Price Per Unit (Harga per Lembar/Unit)** sebagai angka mutlak.
    *   Total nilai transaksi (`Total Value`) dihitung otomatis:
        *   Tipe *Buy*: `(Quantity * Price) + Fee`
        *   Tipe *Sell*: `(Quantity * Price) - Fee`
    *   Pengurangan unit saat *Sell* dihitung otomatis secara matematis oleh backend (tanpa input nominal negatif).
*   **Kalkulasi Untung/Rugi & Kinerja Aset**:
    *   *Harga Beli Rata-rata (Average Price)*: Total nilai beli bersih dibagi dengan total unit/lembar saat ini.
    *   *Harga Pasar Saat Ini (Current Price)*:
        *   **Saham**: Diambil otomatis secara online dari Yahoo Finance API oleh backend Go.
        *   **Emas**: Di-scrape otomatis oleh backend Go dari situs harga emas lokal tepercaya.
        *   **Reksa Dana (RDPU/RDPT)**: Disediakan **Fitur Quick-Update NAV** di UI. Pengguna dapat memperbarui NAV per unit dengan cepat melalui input box kecil di samping nama aset, yang langsung memperbarui nilai pasar aset tersebut secara offline.
    *   *Floating Gain/Loss*: Selisih nilai pasar terkini dengan total nilai investasi dalam nominal Rupiah dan persentase (%).
*   **Deteksi Konsentrasi Risiko**: Menampilkan porsi persentase suatu saham individu terhadap total portofolio saham untuk mencegah eksposur tunggal yang berlebihan (misal: porsi BBRI atau ASII).

### E. Asisten Evaluasi Finansial (Check-in Assistant)
*   **Laporan Ringkasan Bulanan/Mingguan**: Tombol sekali klik untuk menganalisis pengeluaran dalam rentang tanggal tertentu, mengelompokkan kategori pengeluaran terbesar, dan mendeteksi pemborosan.
*   **Deteksi Konsentrasi Aset**: Logika cerdas untuk menganalisis eksposur risiko (misal: memberikan peringatan jika porsi satu saham individu melebihi 30% dari total portofolio saham, menyarankan alokasi DCA berikutnya ke aset lain).

### F. Fitur Impor & Migrasi Data dari Notion (Data Import & Migration)
*   **Notion API Migrator**: Script backend Golang yang memanfaatkan kredensial integrasi Notion untuk membaca database historis Anda (Income, Expenses, Budget, Portfolio Asset, Journal Entries) dan menulisnya ke file lokal CSV dan JSON.
*   **Pembersihan Data Saham Pecahan**: Script migrasi otomatis membulatkan total unit saham ke angka lot terdekat atau menyesuaikan harga beli rata-rata agar data historis saham tidak lagi berbentuk pecahan.

---

## 4. Kebutuhan Non-Fungsional (Non-Functional Requirements)

*   **Penyimpanan Data Lokal**: Data disimpan dalam format CSV terbuka (`income_v2.csv`, `expenses_v2.csv`, `budget_v3.csv`) dan JSON (`portfolio.json`) agar mudah di-backup atau dipindahkan kembali ke Notion/Excel jika diperlukan.
*   **Auto-Backup**: Aplikasi akan menduplikasi file CSV dan JSON ke folder `backup/` berformat timestamp setiap kali server lokal backend Golang dinyalakan.
*   **Performa Ekstrim**: Waktu pemuatan halaman di browser lokal harus di bawah **200ms**.
*   **Aestetika Glassmorphism Dark Mode**: Antarmuka modern bertema gelap menggunakan Vanilla CSS murni dengan variabel HSL yang harmonis, efek blur glassmorphic (`backdrop-filter`), dan mikro-animasi pada tombol/kartu.
*   **Keamanan**: Berjalan sepenuhnya di `localhost` tanpa memerlukan koneksi internet, kecuali saat memanggil API luar untuk memperbarui harga saham/emas terkini.

---

## 5. Rencana Rilis & Fase Pengembangan (Milestones)

### Fase 1: Fondasi Backend & Database Lokal (Ideal Schema)
*   Setup server Golang lokal dan integrasi pembaca/penulis file CSV & JSON secara thread-safe menggunakan mutex.
*   Inisialisasi file data lokal dengan data sampel berdasarkan skema ideal.
*   Implementasi fitur Auto-Backup pada startup backend.

### Fase 2: Implementasi UI & Grafik Interaktif
*   Pembuatan file `index.html` dan CSS bertema Glassmorphic Dark Mode.
*   Penerapan navigasi form input cashflow keyboard-first.
*   Integrasi Chart.js untuk alokasi aset dan kontrol anggaran.
*   Pembuatan modul pembaruan cepat (Quick-Update NAV) di UI portofolio reksa dana.

### Fase 3: Integrasi API Harga Saham & Emas Scraper
*   Integrasi parser Yahoo Finance API untuk harga saham.
*   Pembuatan scraper harga emas lokal otomatis.
*   Implementasi logika perhitungan floating profit/loss dan deteksi konsentrasi risiko.

### Fase 4: Script Migrasi Notion (Fase 0 yang Digeser)
*   Pembuatan script `notion_migrator.go` menggunakan Notion API.
*   Konversi data historis Notion lama ke struktur CSV dan JSON baru yang ideal (termasuk membulatkan unit pecahan saham).

### Fase 5: Asisten Evaluasi Finansial
*   Pengembangan modul Check-in Assistant untuk analisis mingguan/bulanan otomatis.
