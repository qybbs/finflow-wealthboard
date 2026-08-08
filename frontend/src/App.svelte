<script>
  import { onMount, tick } from 'svelte';
  import Chart from 'chart.js/auto';

  // --- STATE ---
  let activeTab = 'dashboard';
  let activeFilter = 'ALL';
  
  let expenses = [];
  let incomes = [];
  let portfolioData = null;
  let accountsData = [];
  let budgetsData = [];
  
  // Charts
  let portfolioCanvas;
  let portfolioChart = null;
  let savingsCanvas;
  let savingsChart = null;
  let incDonutCanvas;
  let incDonutChart = null;
  let expDonutCanvas;
  let expDonutChart = null;
  
  // Forms
  let txType = 'EXPENSE';
  let txCategory = '';
  let txAmount = '';
  let txMethod = '';
  let txDate = new Date().toISOString().split('T')[0];
  
  // Portfolio Form State
  let portTxType = 'BUY';
  let portAssetID = '';
  let portAssetType = 'SAHAM';
  let portAssetCode = '';
  let portQuantity = '';
  let portPrice = '';
  let portFee = '';
  let portMethod = '';
  let portDate = new Date().toISOString().split('T')[0];
  
  let newBudgetCategory = '';
  let newBudgetLimit = '';
  let newBudgetInterval = 'MONTHLY';

  // Derived
  $: uniqueMethods = accountsData.map(a => a.name);
  $: totalPortfolioValue = portfolioData?.assets ? portfolioData.assets.reduce((sum, a) => sum + a.total_value, 0) : 0;
  $: totalFloatingPnL = portfolioData?.assets ? portfolioData.assets.reduce((sum, a) => sum + a.profit_loss, 0) : 0;
  
  let monthlyData = Array.from({length: 12}, (_, i) => ({ month: i+1, income: 0, expense: 0, net: 0 }));

  // --- LIFECYCLE ---
  onMount(async () => {
    await fetchAllData();
  });

  // --- FETCHING ---
  async function fetchAllData() {
    await Promise.all([
        fetchExpenses(),
        fetchIncomes(),
        fetchPortfolio(),
        fetchAccounts(),
        fetchBudgets()
    ]);
    calculateMonthlyData();
    if (activeTab === 'dashboard') {
        setTimeout(renderDashboardCharts, 100);
    }
  }

  async function fetchExpenses() { const res = await fetch('/api/expenses'); expenses = await res.json() || []; }
  async function fetchIncomes() { const res = await fetch('/api/incomes'); incomes = await res.json() || []; }
  async function fetchPortfolio() { const res = await fetch('/api/portfolio'); portfolioData = await res.json(); }
  async function fetchAccounts() { const res = await fetch('/api/accounts'); accountsData = await res.json() || []; }
  async function fetchBudgets() { const res = await fetch('/api/budgets'); budgetsData = await res.json() || []; }

  function calculateMonthlyData() {
      // Reset
      monthlyData = Array.from({length: 12}, (_, i) => ({ month: i+1, income: 0, expense: 0, net: 0 }));
      
      incomes.forEach(inc => {
          const m = parseInt(inc.date.split('-')[1]);
          if(m >= 1 && m <= 12) monthlyData[m-1].income += inc.amount;
      });
      expenses.forEach(exp => {
          const m = parseInt(exp.date.split('-')[1]);
          if(m >= 1 && m <= 12) monthlyData[m-1].expense += exp.amount;
      });
      monthlyData.forEach(d => d.net = d.income - d.expense);
  }

  // --- ACTIONS ---
  async function submitTransaction() {
    if (!txCategory || !txAmount || !txMethod) return;
    const newTx = {
      id: "tx_" + Date.now(),
      date: txDate,
      type: txType,
      category: txCategory,
      amount: parseFloat(txAmount),
      description: "",
      method: txMethod.toUpperCase()
    };
    try {
      const endpoint = txType === 'INCOME' ? '/api/incomes/add' : '/api/expenses/add';
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newTx)
      });
      if (res.ok) {
        txCategory = ''; txAmount = ''; txMethod = '';
        await fetchAllData(); // Refresh all
      }
    } catch (err) {}
  }

  async function submitPortfolioTx() {
      if(!portAssetCode || !portQuantity || !portPrice || !portMethod) return;
      const req = {
          id: "ptx_" + Date.now(),
          date: portDate,
          asset_id: "asset_" + portAssetCode.toUpperCase(),
          asset_type: portAssetType,
          asset_code: portAssetCode.toUpperCase(),
          type: portTxType,
          quantity: parseFloat(portQuantity),
          price_per_unit: parseFloat(portPrice),
          fee: parseFloat(portFee) || 0,
          method: portMethod.toUpperCase()
      };
      
      try {
          const res = await fetch('/api/portfolio/transaction', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(req)
          });
          if(res.ok) {
              portAssetCode = ''; portQuantity = ''; portPrice = ''; portFee = ''; portMethod = '';
              await fetchAllData();
          }
      } catch (err) {}
  }

  async function updateNAV(assetID, currentPrice) {
      const newPrice = prompt("Masukkan NAV / Harga baru:", currentPrice);
      if(newPrice === null || isNaN(newPrice)) return;
      
      try {
          const res = await fetch('/api/portfolio/update-price', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ asset_id: assetID, price: parseFloat(newPrice) })
          });
          if(res.ok) {
              await fetchAllData();
          }
      } catch (err) {}
  }

  async function submitBudget() {
    if(!newBudgetCategory || !newBudgetLimit) return;
    
    // Check if category exists
    let exists = budgetsData.find(b => b.category === newBudgetCategory);
    let payload = [...budgetsData];
    if(exists) {
        exists.limit = parseFloat(newBudgetLimit);
        exists.interval = newBudgetInterval;
    } else {
        payload.push({ category: newBudgetCategory, limit: parseFloat(newBudgetLimit), interval: newBudgetInterval });
    }

    try {
      const res = await fetch('/api/budgets/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      if (res.ok) {
        newBudgetCategory = ''; newBudgetLimit = '';
        await fetchBudgets();
      }
    } catch (err) {}
  }

  function handleKeydown(event) {
    if (event.key === 'Enter') submitTransaction();
  }

  // --- CHARTS ---
  async function setTab(tab) {
      activeTab = tab;
      await tick();
      if (tab === 'dashboard') {
          renderDashboardCharts();
      } else if (tab === 'portfolio') {
          renderPortfolioChart();
      }
  }

  function renderPortfolioChart() {
    if (!portfolioCanvas || !portfolioData || !portfolioData.assets) return;
    if (portfolioChart) portfolioChart.destroy();

    portfolioChart = new Chart(portfolioCanvas, {
      type: 'doughnut',
      data: {
        labels: portfolioData.assets.map(a => a.code),
        datasets: [{
          data: portfolioData.assets.map(a => a.total_value),
          backgroundColor: ['#38bdf8', '#818cf8', '#34d399', '#fbbf24', '#f87171'],
          borderWidth: 0
        }]
      },
      options: { plugins: { legend: { labels: { color: 'white' } } } }
    });
  }

  function renderDashboardCharts() {
      if(!savingsCanvas || !incDonutCanvas || !expDonutCanvas) return;

      if(savingsChart) savingsChart.destroy();
      savingsChart = new Chart(savingsCanvas, {
          type: 'line',
          data: {
              labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'],
              datasets: [{
                  label: 'Net Savings (Rp)',
                  data: monthlyData.map(d => d.net),
                  borderColor: '#38bdf8',
                  backgroundColor: 'rgba(56, 189, 248, 0.2)',
                  fill: true,
                  tension: 0.4
              }]
          },
          options: {
              plugins: { legend: { labels: { color: 'white' } } },
              scales: {
                  x: { ticks: { color: '#94a3b8' } },
                  y: { ticks: { color: '#94a3b8' } }
              }
          }
      });

      // Income Donut Grouped
      let incMap = {};
      incomes.forEach(i => incMap[i.category] = (incMap[i.category] || 0) + i.amount);
      if(incDonutChart) incDonutChart.destroy();
      incDonutChart = new Chart(incDonutCanvas, {
          type: 'doughnut',
          data: {
              labels: Object.keys(incMap).length > 0 ? Object.keys(incMap) : ['No Data'],
              datasets: [{
                  data: Object.keys(incMap).length > 0 ? Object.values(incMap) : [1],
                  backgroundColor: ['#34d399', '#10b981', '#059669'], borderWidth: 0
              }]
          },
          options: { plugins: { legend: { display: false } }, cutout: '70%' }
      });

      // Expense Donut Grouped
      let expMap = {};
      expenses.forEach(e => expMap[e.category] = (expMap[e.category] || 0) + e.amount);
      if(expDonutChart) expDonutChart.destroy();
      expDonutChart = new Chart(expDonutCanvas, {
          type: 'doughnut',
          data: {
              labels: Object.keys(expMap).length > 0 ? Object.keys(expMap) : ['No Data'],
              datasets: [{
                  data: Object.keys(expMap).length > 0 ? Object.values(expMap) : [1],
                  backgroundColor: ['#f87171', '#ef4444', '#dc2626', '#b91c1c'], borderWidth: 0
              }]
          },
          options: { plugins: { legend: { display: false } }, cutout: '70%' }
      });
  }
  
  // Helpers
  $: filteredExpenses = expenses.filter(tx => {
      if(activeFilter === 'ALL') return true;
      const m = parseInt(tx.date.split('-')[1]);
      if(activeFilter === 'Q1') return m >= 1 && m <= 3;
      if(activeFilter === 'Q2') return m >= 4 && m <= 6;
      if(activeFilter === 'Q3') return m >= 7 && m <= 9;
      if(activeFilter === 'Q4') return m >= 10 && m <= 12;
      return true;
  });
</script>

<main>
  <h1>FinFlow Wealthboard</h1>
  
  <div class="tabs">
    <button class:active={activeTab === 'dashboard'} on:click={() => setTab('dashboard')}>Dashboard</button>
    <button class:active={activeTab === 'cashflow'} on:click={() => setTab('cashflow')}>Cashflow</button>
    <button class:active={activeTab === 'portfolio'} on:click={() => setTab('portfolio')}>Portfolio</button>
  </div>
  
  <div class="container fade-in">
    {#if activeTab === 'dashboard'}
       <div class="row">
           <div class="card glass-panel flex-2">
               <h2>Savings Trend (2026)</h2>
               <canvas bind:this={savingsCanvas}></canvas>
           </div>
           <div class="card glass-panel flex-1 center-content">
               <h2>Income Breakdown</h2>
               <div class="donut-wrapper"><canvas bind:this={incDonutCanvas}></canvas></div>
           </div>
           <div class="card glass-panel flex-1 center-content">
               <h2>Expense Breakdown</h2>
               <div class="donut-wrapper"><canvas bind:this={expDonutCanvas}></canvas></div>
           </div>
       </div>
       <h2 style="margin-top: 30px;">Total Saving (Monthly)</h2>
       <div class="grid-12">
           {#each monthlyData as data}
           <div class="month-card glass-panel">
               <h4>Bulan {data.month}</h4>
               <div class="stat"><span class="text-green">In:</span> Rp {data.income.toLocaleString('id-ID')}</div>
               <div class="stat"><span class="text-red">Out:</span> Rp {data.expense.toLocaleString('id-ID')}</div>
               <div class="stat-net" style="color: {data.net >= 0 ? '#34d399' : '#f87171'}">
                   Net: Rp {data.net.toLocaleString('id-ID')}
               </div>
           </div>
           {/each}
       </div>

    {:else if activeTab === 'cashflow'}
       <div class="row">
           <!-- LEFT COLUMN -->
           <div class="col flex-1">
               <div class="card glass-panel">
                   <h2>Saldo Rekening</h2>
                   <table class="minimal-table">
                       <tbody>
                           {#each accountsData as acc}
                           <tr>
                               <td>{acc.name}</td>
                               <td class="amount" style="color: {acc.balance >= 0 ? '#38bdf8' : '#f87171'}">
                                   Rp {acc.balance.toLocaleString('id-ID')}
                               </td>
                           </tr>
                           {/each}
                       </tbody>
                   </table>
               </div>
               
               <div class="card glass-panel" style="margin-top: 20px;">
                   <div style="display: flex; justify-content: space-between; align-items: center;">
                       <h2>Anggaran (Budget)</h2>
                   </div>
                   <!-- Set budget inline form -->
                   <div class="budget-form">
                       <input type="text" bind:value={newBudgetCategory} placeholder="Kategori" />
                       <input type="number" bind:value={newBudgetLimit} placeholder="Limit (Rp)" />
                       <select bind:value={newBudgetInterval}>
                           <option value="WEEKLY">Mingguan</option>
                           <option value="MONTHLY">Bulanan</option>
                       </select>
                       <button class="btn-small" on:click={submitBudget}>Set</button>
                   </div>
                   
                   {#each budgetsData as b}
                   <div class="budget-item">
                       <div class="budget-head">
                           <strong>{b.category}</strong> <span class="tag">{b.interval}</span>
                       </div>
                       <div class="budget-bar-bg">
                           <!-- Math to cap at 100% -->
                           <div class="budget-bar-fill" 
                                style="width: {Math.min((b.spent / b.limit) * 100, 100)}%; 
                                       background-color: {(b.spent / b.limit) > 0.9 ? '#ef4444' : (b.spent / b.limit) > 0.7 ? '#facc15' : '#10b981'}">
                           </div>
                       </div>
                       <div class="budget-details">
                           <span>Spent: Rp {b.spent.toLocaleString('id-ID')}</span>
                           <span>Rem: Rp {b.remaining.toLocaleString('id-ID')}</span>
                       </div>
                   </div>
                   {/each}
               </div>
           </div>
           
           <!-- RIGHT COLUMN -->
           <div class="col flex-2">
               <div class="card glass-panel">
                   <h2>Tambah Transaksi Cepat</h2>
                   <div style="margin-bottom: 10px; font-size: 0.9rem;">
                       <label style="margin-right: 15px;"><input type="radio" bind:group={txType} value="EXPENSE"> Pengeluaran</label>
                       <label><input type="radio" bind:group={txType} value="INCOME"> Pemasukan</label>
                   </div>
                   <div class="form-row">
                       <input type="date" bind:value={txDate} />
                       
                       <input type="text" list="methods-list" bind:value={txMethod} placeholder="Metode/Rekening (Bebas)" />
                       <datalist id="methods-list">
                           {#each uniqueMethods as m} <option value={m}></option> {/each}
                       </datalist>

                       <input type="text" bind:value={txCategory} placeholder="Kategori" />
                       <input type="number" bind:value={txAmount} on:keydown={handleKeydown} placeholder="Nominal (Rp)" />
                       <button class="btn-primary" on:click={submitTransaction}>Save</button>
                   </div>
               </div>
               
               <div class="card glass-panel" style="margin-top: 20px;">
                   <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                       <h2>Histori Transaksi</h2>
                       <div class="filter-group">
                           <button class:active={activeFilter==='ALL'} on:click={() => activeFilter='ALL'}>ALL</button>
                           <button class:active={activeFilter==='Q1'} on:click={() => activeFilter='Q1'}>Q1</button>
                           <button class:active={activeFilter==='Q2'} on:click={() => activeFilter='Q2'}>Q2</button>
                           <button class:active={activeFilter==='Q3'} on:click={() => activeFilter='Q3'}>Q3</button>
                           <button class:active={activeFilter==='Q4'} on:click={() => activeFilter='Q4'}>Q4</button>
                       </div>
                   </div>
                   <div class="table-container">
                       <table class="full-table">
                           <thead><tr><th>Tanggal</th><th>Rekening</th><th>Kategori</th><th>Nominal</th></tr></thead>
                           <tbody>
                               {#each filteredExpenses as tx}
                               <tr>
                                   <td>{tx.date}</td>
                                   <td><span class="tag">{tx.method}</span></td>
                                   <td>{tx.category}</td>
                                   <td class="amount" style="color: {tx.type === 'INCOME' ? '#34d399' : '#f87171'}">
                                       {tx.type === 'INCOME' ? '+' : '-'} Rp {tx.amount.toLocaleString('id-ID')}
                                   </td>
                               </tr>
                               {/each}
                           </tbody>
                       </table>
                   </div>
               </div>
           </div>
       </div>

    {:else if activeTab === 'portfolio'}
       <div class="row">
           <div class="card glass-panel flex-1 center-content">
             <h2>Alokasi Aset</h2>
             <div class="chart-container"><canvas bind:this={portfolioCanvas}></canvas></div>
           </div>
           
           <div class="card glass-panel flex-2">
             <h2>Mutasi Portofolio & Aset</h2>
             <div class="form-row" style="margin-bottom: 20px;">
                 <select bind:value={portTxType} style="width: auto;">
                     <option value="BUY">BUY</option>
                     <option value="SELL">SELL</option>
                     <option value="DIVIDEND">DIVIDEND</option>
                 </select>
                 <select bind:value={portAssetType} style="width: auto;">
                     <option value="SAHAM">SAHAM</option>
                     <option value="REKSA_DANA">REKSA DANA</option>
                     <option value="EMAS">EMAS</option>
                 </select>
                 <input type="text" bind:value={portAssetCode} placeholder="Kode Aset" />
                 <input type="number" bind:value={portQuantity} placeholder="Jumlah Unit" />
                 <input type="number" bind:value={portPrice} placeholder="Harga/Unit (Rp)" />
             </div>
             <div class="form-row" style="margin-bottom: 30px;">
                 <input type="number" bind:value={portFee} placeholder="Fee Transaksi (Rp)" />
                 <input type="text" list="methods-list" bind:value={portMethod} placeholder="Metode/Rekening Pembayaran" />
                 <input type="date" bind:value={portDate} style="width: 150px;"/>
                 <button class="btn-primary" on:click={submitPortfolioTx}>Proses Transaksi</button>
             </div>

             <div class="flex-between">
                 <h2>Daftar Aset</h2>
                 <div style="text-align: right;">
                     <div>Total Nilai: <strong class="text-blue">Rp {totalPortfolioValue.toLocaleString('id-ID')}</strong></div>
                     <div>Floating PnL: <strong class={totalFloatingPnL >= 0 ? 'text-green' : 'text-red'}>Rp {totalFloatingPnL.toLocaleString('id-ID')}</strong></div>
                 </div>
             </div>
             <table class="full-table" style="margin-top: 15px;">
                 <thead><tr><th>Kode</th><th>Tipe</th><th>Unit</th><th>Market Price</th><th>Total Value</th><th>PnL (%)</th></tr></thead>
                 <tbody>
                     {#if portfolioData && portfolioData.assets}
                         {#each portfolioData.assets as asset}
                         <tr>
                             <td>
                                {asset.code}
                                {#if asset.profit_loss_pct <= -10} <span title="Alert!" style="color: #fbbf24;">⚠️</span> {/if}
                             </td>
                             <td>
                                {asset.type}
                                {#if asset.type === 'REKSA_DANA'}
                                   <button class="btn-small" style="margin-left: 5px; padding: 2px 5px;" on:click={() => updateNAV(asset.id, asset.current_price)}>✏️</button>
                                {/if}
                             </td>
                             <td>{asset.quantity}</td>
                             <td>Rp {asset.current_price.toLocaleString('id-ID')}</td>
                             <td>Rp {asset.total_value.toLocaleString('id-ID')}</td>
                             <td class={asset.profit_loss >= 0 ? 'text-green' : 'text-red'}>{asset.profit_loss_pct.toFixed(2)}%</td>
                         </tr>
                         {/each}
                     {/if}
                 </tbody>
             </table>
           </div>
       </div>
    {/if}
  </div>
</main>

<style>
  :global(body) { background-color: #0f172a; color: #f8fafc; font-family: 'Inter', sans-serif; margin: 0; padding: 30px; }
  h1 { text-align: center; background: linear-gradient(to right, #38bdf8, #818cf8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin-bottom: 20px; font-size: 2rem; }
  h2 { margin-top: 0; font-size: 1.1rem; color: #e2e8f0; font-weight: 600; }
  
  .tabs { display: flex; justify-content: center; gap: 15px; margin-bottom: 30px; }
  .tabs button { background: rgba(255,255,255,0.05); color: #94a3b8; border: 1px solid rgba(255,255,255,0.1); padding: 8px 24px; border-radius: 20px; cursor: pointer; transition: 0.3s; font-weight: bold; }
  .tabs button.active { background: rgba(56, 189, 248, 0.15); color: #38bdf8; border-color: #38bdf8; }
  
  .container { max-width: 1200px; margin: 0 auto; }
  .row { display: flex; gap: 20px; flex-wrap: wrap; }
  .col { display: flex; flex-direction: column; gap: 20px; }
  .flex-1 { flex: 1; } .flex-2 { flex: 2; }
  .flex-between { display: flex; justify-content: space-between; align-items: center; }
  
  .glass-panel { background: rgba(255,255,255,0.03); backdrop-filter: blur(10px); -webkit-backdrop-filter: blur(10px); border: 1px solid rgba(255,255,255,0.1); border-radius: 12px; padding: 25px; box-shadow: 0 4px 20px rgba(0,0,0,0.2); }
  
  /* Utilities */
  .center-content { display: flex; flex-direction: column; align-items: center; justify-content: center; }
  .text-green { color: #34d399; } .text-red { color: #f87171; } .text-blue { color: #38bdf8; }
  .fade-in { animation: fadeIn 0.3s ease-in; }
  @keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
  
  /* Charts */
  .donut-wrapper { width: 160px; height: 160px; margin-top: 10px; }
  .chart-container { max-width: 250px; margin: 0 auto; }
  
  /* Grid 12 Months */
  .grid-12 { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; }
  .month-card { padding: 15px; text-align: center; }
  .month-card h4 { margin: 0 0 10px 0; color: #94a3b8; font-size: 0.9rem; }
  .month-card .stat { font-size: 0.85rem; margin-bottom: 5px; }
  .month-card .stat-net { margin-top: 10px; font-weight: bold; border-top: 1px solid rgba(255,255,255,0.1); padding-top: 5px; }
  
  /* Forms */
  input, select { padding: 10px 12px; border-radius: 6px; border: 1px solid rgba(255,255,255,0.2); background: rgba(0,0,0,0.3); color: white; font-size: 0.9rem; width: 100%; box-sizing: border-box; }
  input:focus, select:focus { outline: none; border-color: #38bdf8; }
  .form-row { display: flex; gap: 10px; margin-top: 15px; align-items: center; }
  
  .btn-primary { background: #38bdf8; color: #0f172a; font-weight: bold; padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer; transition: 0.2s; white-space: nowrap; }
  .btn-primary:hover { background: #7dd3fc; }
  .btn-small { background: rgba(56, 189, 248, 0.2); color: #38bdf8; border: 1px solid #38bdf8; border-radius: 4px; padding: 5px 10px; cursor: pointer; }
  
  /* Tables */
  .minimal-table { width: 100%; border-collapse: collapse; }
  .minimal-table td { padding: 8px 0; border-bottom: 1px solid rgba(255,255,255,0.05); }
  .minimal-table .amount { text-align: right; font-weight: bold; }
  
  .full-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  .full-table th, .full-table td { padding: 12px 10px; border-bottom: 1px solid rgba(255,255,255,0.05); text-align: left; }
  .full-table th { color: #94a3b8; font-weight: 500; }
  .full-table .amount { text-align: right; font-family: monospace; font-size: 1rem; }
  .table-container { max-height: 400px; overflow-y: auto; }
  
  /* Filter */
  .filter-group { display: flex; gap: 5px; }
  .filter-group button { background: none; border: 1px solid rgba(255,255,255,0.2); color: #94a3b8; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 0.8rem; }
  .filter-group button.active { background: #38bdf8; color: #0f172a; border-color: #38bdf8; font-weight: bold; }
  
  /* Budget Tracker */
  .budget-form { display: grid; grid-template-columns: 1fr 1fr 1fr auto; gap: 10px; margin-bottom: 20px; }
  .budget-item { margin-bottom: 15px; }
  .budget-head { display: flex; justify-content: space-between; font-size: 0.9rem; margin-bottom: 5px; }
  .budget-bar-bg { width: 100%; height: 6px; background: rgba(255,255,255,0.1); border-radius: 3px; overflow: hidden; margin-bottom: 5px; }
  .budget-bar-fill { height: 100%; transition: width 0.3s; }
  .budget-details { display: flex; justify-content: space-between; font-size: 0.8rem; color: #94a3b8; }
  
  .tag { background: rgba(255,255,255,0.1); padding: 2px 6px; border-radius: 4px; font-size: 0.75rem; }
</style>
