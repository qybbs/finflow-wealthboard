<script>
  import { onMount } from 'svelte';
  import Chart from 'chart.js/auto';

  // 1. STATE VARIABLES
  // Svelte otomatis merender ulang layar ketika variabel ini berubah!
  let expenses = [];
  let portfolioCanvas;
  let portfolioData = null;
  
  // Variabel untuk Form (Two-Way Binding)
  let txCategory = '';
  let txAmount = '';

  // 2. LIFECYCLE HOOK
  onMount(async () => {
    fetchExpenses();
    fetchPortfolio();
  });

  async function fetchExpenses() {
    try {
      const res = await fetch('/api/expenses');
      expenses = await res.json();
    } catch (err) {
      console.error("Gagal mengambil data:", err);
    }
  }

  async function fetchPortfolio() {
    try {
      const res = await fetch('/api/portfolio');
      portfolioData = await res.json();
      renderChart(); // Gambar grafiknya setelah data dapat!
    } catch (err) {
      console.error("Gagal mengambil data portfolio:", err);
    }
  }

  // 3. FUNGSI SUBMIT FORM
  async function submitTransaction() {
    if (!txCategory || !txAmount) return; // Validasi sederhana

    const newTx = {
      id: "tx_" + Date.now(), // Generate ID acak berdasarkan waktu
      date: new Date().toISOString().split('T')[0], // YYYY-MM-DD
      type: "EXPENSE",
      category: txCategory,
      amount: parseFloat(txAmount),
      description: "",
      method: 'CASH'
    };

    try {
      const res = await fetch('/api/expenses/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newTx)
      });

      if (res.ok) {
        // Reset form
        txCategory = '';
        txAmount = '';
        // Fetch ulang data dari Go
        fetchExpenses();
      }
    } catch (err) {
      console.error("Gagal menyimpan:", err);
    }
  }

  function renderChart() {
    // Pastikan canvas dan datanya tidak kosong
    if (!portfolioCanvas || !portfolioData || !portfolioData.assets) return;
    // Menyiapkan label (contoh: ["BBCA", "GOTO"]) dan data (contoh: [1000000, 500000])
    const labels = portfolioData.assets.map(asset => asset.code);
    const dataValues = portfolioData.assets.map(asset => asset.total_value);
    // Membuat Pie Chart
    new Chart(portfolioCanvas, {
      type: 'doughnut', // Jenis grafik donat
      data: {
        labels: labels,
        datasets: [{
          label: 'Total Value (Rp)',
          data: dataValues,
          backgroundColor: ['#38bdf8', '#818cf8', '#34d399', '#fbbf24', '#f87171'],
          borderWidth: 0,
          hoverOffset: 10
        }]
      },
      options: {
        plugins: {
          legend: {
            labels: { color: 'white' } // Teks legend warna putih untuk dark mode
          }
        }
      }
    });
  }

  // 4. EVENT LISTENER KEYBOARD
  function handleKeydown(event) {
    if (event.key === 'Enter') {
      submitTransaction();
    }
  }
</script>

<main>
  <h1>FinFlow Wealthboard</h1>
  
  <div class="dashboard">
    <!-- Kolom Kiri: Form Input -->
    <div class="card glass-panel">
      <h2>Tambah Pengeluaran Cepat</h2>
      
      <div class="form-group">
        <label>Kategori</label>
        <!-- AJAIB 1: bind:value menyambungkan input ini langsung ke txCategory -->
        <input type="text" bind:value={txCategory} placeholder="Contoh: Makan Siang" />
      </div>
      
      <div class="form-group">
        <label>Nominal (Rp)</label>
        <!-- AJAIB 2: on:keydown menangkap tombol Enter -->
        <input type="number" bind:value={txAmount} on:keydown={handleKeydown} placeholder="50000" />
      </div>
      
      <button on:click={submitTransaction}>Simpan (Tekan Enter)</button>
    </div>

    <!-- Kolom Kanan: Tabel Data -->
    <div class="card glass-panel">
      <h2>Riwayat Pengeluaran</h2>
      <table>
        <thead>
          <tr>
            <th>Kategori</th>
            <th>Nominal</th>
          </tr>
        </thead>
        <tbody>
          <!-- AJAIB 3: {#each} untuk looping data -->
          {#each expenses as tx}
            <tr>
              <td>{tx.category}</td>
              <td class="amount">Rp {tx.amount.toLocaleString('id-ID')}</td>
            </tr>
          {/each}
          
          <!-- AJAIB 4: {#if} untuk mengecek array kosong -->
          {#if expenses.length === 0}
            <tr><td colspan="2" style="text-align: center; color: #888;">Belum ada pengeluaran tersimpan.</td></tr>
          {/if}
        </tbody>
      </table>
    </div>

        <!-- Kolom Ketiga: Grafik Portofolio -->
    <div class="card glass-panel">
      <h2>Alokasi Portofolio</h2>
      
      <!-- AJAIB 5: bind:this menyambungkan elemen canvas ini ke variabel JS kita -->
      <canvas bind:this={portfolioCanvas}></canvas>
      
      {#if !portfolioData || portfolioData.assets.length === 0}
        <p style="text-align: center; color: #888; margin-top: 20px;">Belum ada aset terdaftar.</p>
      {/if}
    </div>

  </div>
</main>

<style>
  /* Mengatur background keseluruhan (global modifier svelte) */
  :global(body) {
    background-color: #0f172a; /* Warna biru super gelap (dark mode) */
    color: #f8fafc;
    font-family: 'Inter', sans-serif;
    margin: 0;
    padding: 40px;
  }

  h1 {
    text-align: center;
    background: linear-gradient(to right, #38bdf8, #818cf8);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    margin-bottom: 40px;
  }

  .dashboard {
    display: flex;
    gap: 30px;
    max-width: 1000px;
    margin: 0 auto;
  }

  .card {
    flex: 1;
  }

  /* --- DESAIN GLASSMORPHISM UTAMA --- */
  .glass-panel {
    background: rgba(255, 255, 255, 0.05); /* Agak transparan */
    backdrop-filter: blur(12px); /* Efek kaca buram */
    -webkit-backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    padding: 30px;
    box-shadow: 0 4px 30px rgba(0, 0, 0, 0.2);
  }

  h2 {
    margin-top: 0;
    font-size: 1.2rem;
    color: #e2e8f0;
  }

  .form-group {
    margin-bottom: 20px;
  }

  label {
    display: block;
    margin-bottom: 8px;
    font-size: 0.9rem;
    color: #cbd5e1;
  }

  input {
    width: 90%;
    padding: 12px 15px;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(0, 0, 0, 0.2);
    color: white;
    font-size: 1rem;
    transition: all 0.3s;
  }

  input:focus {
    outline: none;
    border-color: #38bdf8;
    background: rgba(0, 0, 0, 0.4);
  }

  button {
    width: 100%;
    background: #38bdf8;
    color: #0f172a;
    font-weight: bold;
    padding: 12px;
    border: none;
    border-radius: 8px;
    cursor: pointer;
    font-size: 1rem;
    transition: transform 0.1s, background 0.3s;
  }

  button:hover {
    background: #7dd3fc;
  }

  button:active {
    transform: scale(0.98);
  }

  /* Tabel CSS */
  table {
    width: 100%;
    border-collapse: collapse;
  }
  
  th, td {
    padding: 12px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  th {
    text-align: left;
    color: #94a3b8;
    font-weight: 500;
  }
  
  .amount {
    text-align: right;
    font-family: monospace;
    font-size: 1.1rem;
    color: #ef4444; /* Merah untuk pengeluaran */
  }
</style>
