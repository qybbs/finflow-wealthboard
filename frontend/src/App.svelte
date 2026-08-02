<script>
  import { onMount } from 'svelte';
  import Chart from 'chart.js/auto';

  // --- STATE ---
  let activeTab = 'portfolio';
  
  let expenses = [];
  let portfolioData = null;
  let analyticsData = null;
  let portfolioCanvas;
  let chartInstance = null;
  
  // Forms
  let txCategory = '';
  let txAmount = '';

  onMount(async () => {
    fetchExpenses();
    fetchPortfolio();
    fetchAnalytics();
  });

  // --- DATA FETCHING ---
  async function fetchExpenses() {
    try {
      const res = await fetch('/api/expenses');
      expenses = await res.json();
    } catch (err) {
      console.error(err);
    }
  }

  async function fetchPortfolio() {
    try {
      const res = await fetch('/api/portfolio');
      portfolioData = await res.json();
      if (activeTab === 'portfolio') {
          setTimeout(renderChart, 100); // Wait for DOM if tab just switched
      }
    } catch (err) {
      console.error(err);
    }
  }

  async function fetchAnalytics() {
    try {
      const res = await fetch('/api/analytics');
      analyticsData = await res.json();
    } catch (err) {
      console.error(err);
    }
  }

  // --- CHART RENDERING ---
  function renderChart() {
    if (!portfolioCanvas || !portfolioData || !portfolioData.assets) return;
    
    if (chartInstance) {
        chartInstance.destroy(); // Destroy previous chart
    }

    const labels = portfolioData.assets.map(asset => asset.code);
    const dataValues = portfolioData.assets.map(asset => asset.total_value);
    
    chartInstance = new Chart(portfolioCanvas, {
      type: 'doughnut',
      data: {
        labels: labels,
        datasets: [{
          label: 'Total Value (Rp)',
          data: dataValues,
          backgroundColor: ['#38bdf8', '#818cf8', '#34d399', '#fbbf24', '#f87171', '#a78bfa'],
          borderWidth: 0,
          hoverOffset: 10
        }]
      },
      options: {
        plugins: {
          legend: { labels: { color: 'white' } }
        }
      }
    });
  }

  // Switch tab wrapper to re-render chart correctly
  function setTab(tab) {
      activeTab = tab;
      if (tab === 'portfolio') {
          setTimeout(renderChart, 100); // DOM needs to render canvas first
      }
  }

  // --- ACTIONS ---
  async function submitTransaction() {
    if (!txCategory || !txAmount) return;
    const newTx = {
      id: "tx_" + Date.now(),
      date: new Date().toISOString().split('T')[0],
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
        txCategory = ''; txAmount = '';
        fetchExpenses();
      }
    } catch (err) {}
  }

  async function updateNav(asset) {
    // Optimistic or just call API
    // We update the entire portfolio struct for simplicity
    try {
      const res = await fetch('/api/portfolio/update', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(portfolioData)
      });
      if (res.ok) {
          fetchPortfolio();
      }
    } catch(err) {}
  }

  function handleKeydown(event) {
    if (event.key === 'Enter') submitTransaction();
  }

  // --- DERIVED (COMPUTED) VALUES ---
  $: totalPortfolioValue = portfolioData?.assets ? portfolioData.assets.reduce((sum, a) => sum + a.total_value, 0) : 0;
  $: totalFloatingPnL = portfolioData?.assets ? portfolioData.assets.reduce((sum, a) => sum + a.profit_loss, 0) : 0;
</script>

<main>
  <h1>FinFlow Wealthboard</h1>
  
  <div class="tabs">
    <button class:active={activeTab === 'cashflow'} on:click={() => setTab('cashflow')}>Cashflow</button>
    <button class:active={activeTab === 'portfolio'} on:click={() => setTab('portfolio')}>Portfolio</button>
    <button class:active={activeTab === 'analytics'} on:click={() => setTab('analytics')}>Analytics</button>
  </div>
  
  <div class="dashboard">
    {#if activeTab === 'cashflow'}
      <div class="card glass-panel fade-in">
        <h2>Tambah Pengeluaran</h2>
        <div class="form-group">
          <label>Kategori</label>
          <input type="text" bind:value={txCategory} placeholder="Contoh: Makan Siang" />
        </div>
        <div class="form-group">
          <label>Nominal (Rp)</label>
          <input type="number" bind:value={txAmount} on:keydown={handleKeydown} placeholder="50000" />
        </div>
        <button class="btn-primary" on:click={submitTransaction}>Simpan (Tekan Enter)</button>
      </div>

      <div class="card glass-panel fade-in">
        <h2>Riwayat Pengeluaran</h2>
        <table>
          <thead><tr><th>Kategori</th><th>Nominal</th></tr></thead>
          <tbody>
            {#each expenses as tx}
              <tr><td>{tx.category}</td><td class="amount">Rp {tx.amount.toLocaleString('id-ID')}</td></tr>
            {/each}
            {#if expenses.length === 0}
              <tr><td colspan="2" style="text-align: center; color: #888;">Belum ada pengeluaran tersimpan.</td></tr>
            {/if}
          </tbody>
        </table>
      </div>

    {:else if activeTab === 'portfolio'}
      <div class="card glass-panel fade-in">
        <h2>Alokasi Aset</h2>
        <div class="chart-container">
          <canvas bind:this={portfolioCanvas}></canvas>
        </div>
        {#if !portfolioData || portfolioData.assets == null || portfolioData.assets.length === 0}
          <p style="text-align: center; color: #888;">Belum ada aset terdaftar.</p>
        {/if}
      </div>

      <div class="card glass-panel fade-in wide-card">
        <div class="portfolio-header">
            <h2>Daftar Aset</h2>
            <div class="portfolio-summary">
                <div>Total Nilai: <strong class="text-blue">Rp {totalPortfolioValue.toLocaleString('id-ID')}</strong></div>
                <div>Floating PnL: <strong class={totalFloatingPnL >= 0 ? 'text-green' : 'text-red'}>Rp {totalFloatingPnL.toLocaleString('id-ID')}</strong></div>
            </div>
        </div>
        <table>
          <thead>
            <tr>
              <th>Kode</th>
              <th>Tipe</th>
              <th>Unit</th>
              <th>Harga Pasar</th>
              <th>Total Nilai</th>
              <th>PnL (%)</th>
              <th>Aksi</th>
            </tr>
          </thead>
          <tbody>
            {#if portfolioData && portfolioData.assets}
                {#each portfolioData.assets as asset}
                <tr>
                    <td>
                        {asset.code}
                        {#if asset.profit_loss_pct <= -10}
                           <span title="Satpam Virtual: Floating Loss di atas 10%!" style="color: #fbbf24; margin-left: 5px;">⚠️</span>
                        {/if}
                    </td>
                    <td>{asset.type}</td>
                    <td>{asset.quantity}</td>
                    <td>
                        {#if asset.type === 'REKSA_DANA'}
                           <input type="number" bind:value={asset.current_price} class="nav-input" />
                        {:else}
                           Rp {asset.current_price.toLocaleString('id-ID')}
                        {/if}
                    </td>
                    <td>Rp {asset.total_value.toLocaleString('id-ID')}</td>
                    <td class={asset.profit_loss >= 0 ? 'text-green' : 'text-red'}>
                        {asset.profit_loss_pct.toFixed(2)}%
                    </td>
                    <td>
                        {#if asset.type === 'REKSA_DANA'}
                            <button class="btn-small" on:click={() => updateNav(asset)}>Update NAV</button>
                        {:else}
                            <span style="color: #888; font-size: 0.8rem;">Auto-Sync</span>
                        {/if}
                    </td>
                </tr>
                {/each}
            {/if}
          </tbody>
        </table>
      </div>

    {:else if activeTab === 'analytics'}
      <div class="card glass-panel fade-in">
         <h2>Asisten Keuangan FinFlow 🤖</h2>
         {#if analyticsData}
             <div class="analytics-metrics">
                 <div class="metric-box">
                     <h4>Savings Rate</h4>
                     <div class="metric-value" style="color: {analyticsData.savings_rate >= 20 ? '#34d399' : '#f87171'}">
                         {analyticsData.savings_rate.toFixed(1)}%
                     </div>
                 </div>
                 <div class="metric-box">
                     <h4>Emergency Fund Run-Rate</h4>
                     <div class="metric-value" style="color: {analyticsData.emergency_run_rate >= 6 ? '#34d399' : '#f87171'}">
                         {analyticsData.emergency_run_rate.toFixed(1)} Bulan
                     </div>
                 </div>
             </div>
             
             <div class="assistant-chat">
                 <div class="chat-bubble">
                     <p>Halo! Saya menganalisis bahwa:</p>
                     <p><em>{analyticsData.recommendation}</em></p>
                 </div>
             </div>
         {:else}
             <div class="coming-soon">
                 <div class="pulse-dot"></div>
                 <span>Asisten sedang mengumpulkan data Anda...</span>
             </div>
         {/if}
      </div>
    {/if}
  </div>
</main>

<style>
  :global(body) { background-color: #0f172a; color: #f8fafc; font-family: 'Inter', sans-serif; margin: 0; padding: 40px; }
  h1 { text-align: center; background: linear-gradient(to right, #38bdf8, #818cf8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin-bottom: 20px; }
  
  .tabs { display: flex; justify-content: center; gap: 15px; margin-bottom: 30px; }
  .tabs button { background: rgba(255,255,255,0.05); color: #94a3b8; border: 1px solid rgba(255,255,255,0.1); padding: 10px 20px; border-radius: 20px; cursor: pointer; transition: all 0.3s; font-weight: bold; }
  .tabs button.active { background: rgba(56, 189, 248, 0.2); color: #38bdf8; border-color: #38bdf8; }
  
  .dashboard { display: flex; gap: 30px; max-width: 1100px; margin: 0 auto; flex-wrap: wrap; align-items: flex-start; }
  .card { flex: 1; min-width: 300px; }
  .wide-card { flex: 2; min-width: 600px; }
  
  .glass-panel { background: rgba(255, 255, 255, 0.05); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); border: 1px solid rgba(255, 255, 255, 0.1); border-radius: 16px; padding: 30px; box-shadow: 0 4px 30px rgba(0, 0, 0, 0.2); }
  
  h2 { margin-top: 0; font-size: 1.2rem; color: #e2e8f0; }
  .portfolio-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
  .portfolio-summary { text-align: right; font-size: 0.9rem; }
  
  .form-group { margin-bottom: 20px; }
  label { display: block; margin-bottom: 8px; font-size: 0.9rem; color: #cbd5e1; }
  input { width: 90%; padding: 12px 15px; border-radius: 8px; border: 1px solid rgba(255, 255, 255, 0.2); background: rgba(0, 0, 0, 0.2); color: white; font-size: 1rem; transition: all 0.3s; }
  input:focus { outline: none; border-color: #38bdf8; background: rgba(0, 0, 0, 0.4); }
  
  .nav-input { width: 100px; padding: 5px; font-size: 0.9rem; }
  
  .btn-primary { width: 100%; background: #38bdf8; color: #0f172a; font-weight: bold; padding: 12px; border: none; border-radius: 8px; cursor: pointer; transition: transform 0.1s, background 0.3s; }
  .btn-primary:hover { background: #7dd3fc; }
  .btn-primary:active { transform: scale(0.98); }
  
  .btn-small { background: rgba(56, 189, 248, 0.2); color: #38bdf8; border: 1px solid #38bdf8; border-radius: 4px; padding: 5px 10px; cursor: pointer; }
  
  table { width: 100%; border-collapse: collapse; font-size: 0.95rem; }
  th, td { padding: 12px 5px; border-bottom: 1px solid rgba(255, 255, 255, 0.1); }
  th { text-align: left; color: #94a3b8; font-weight: 500; }
  .amount { text-align: right; font-family: monospace; font-size: 1.1rem; color: #ef4444; }
  
  .text-green { color: #34d399; }
  .text-red { color: #f87171; }
  .text-blue { color: #38bdf8; }
  
  .chart-container { max-width: 300px; margin: 0 auto; }
  
  .fade-in { animation: fadeIn 0.4s ease-in-out; }
  @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
  
  .analytics-metrics { display: flex; gap: 20px; margin: 20px 0; }
  .metric-box { flex: 1; background: rgba(0,0,0,0.2); padding: 20px; border-radius: 12px; text-align: center; border: 1px solid rgba(255,255,255,0.05); }
  .metric-box h4 { margin: 0 0 10px 0; color: #cbd5e1; font-weight: normal; font-size: 0.9rem; }
  .metric-value { font-size: 1.8rem; font-weight: bold; }
  
  .assistant-chat { margin-top: 30px; }
  .chat-bubble { background: rgba(56, 189, 248, 0.1); border-left: 4px solid #38bdf8; padding: 15px 20px; border-radius: 0 12px 12px 12px; font-size: 1rem; line-height: 1.5; color: #e2e8f0; }
  
  .coming-soon { display: flex; align-items: center; gap: 10px; margin-top: 20px; color: #94a3b8; }
  .pulse-dot { width: 10px; height: 10px; background-color: #38bdf8; border-radius: 50%; animation: pulse 1.5s infinite; }
  @keyframes pulse { 0% { box-shadow: 0 0 0 0 rgba(56, 189, 248, 0.7); } 70% { box-shadow: 0 0 0 10px rgba(56, 189, 248, 0); } 100% { box-shadow: 0 0 0 0 rgba(56, 189, 248, 0); } }
</style>
