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
  let analyticsData = {};
  
  // Edit State
  let editingTxId = null;
  let editingField = null;
  let editingValue = '';
  
  // New Inline Row State
  let newInlineIncome = {
      date: new Date().toISOString().split('T')[0],
      method: '', category: '', type: 'INCOME', amount: '', description: ''
  };
  let newInlineExpense = {
      date: new Date().toISOString().split('T')[0],
      method: '', category: '', type: 'EXPENSE', amount: '', description: ''
  };

  let expDonutChart = null;

  // Cashflow Charts
  let cfSavingsCanvas;
  let cfSavingsChart = null;
  let cfIncDonutCanvas;
  let cfIncDonutChart = null;
  let cfExpDonutCanvas;
  let cfExpDonutChart = null;

  const categoryColors = ['blue', 'green', 'yellow', 'red', 'purple', 'pink', 'indigo'];
  function getCategoryColor(cat) {
      if (!cat) return 'default';
      const hash = cat.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
      return categoryColors[hash % categoryColors.length];
  }
  
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
    } else if (activeTab === 'cashflow') {
        setTimeout(renderCashflowCharts, 100);
    } else if (activeTab === 'portfolio') {
        setTimeout(renderPortfolioChart, 100);
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

  // --- REACTIVE STATE ---
  $: totalCashBalance = accountsData.reduce((sum, a) => sum + a.balance, 0);
  $: totalPortfolioValue = portfolioData?.assets ? portfolioData.assets.reduce((sum, a) => sum + a.total_value, 0) : 0;
  $: netWorth = totalCashBalance + totalPortfolioValue;

  $: currentMonthSummary = (() => {
      const now = new Date();
      const currentYear = now.getFullYear();
      const currentMonth = now.getMonth() + 1; // 1-12
      
      let income = 0;
      let expense = 0;
      
      incomes.forEach(inc => {
          const parts = inc.date.split('-');
          if (parts.length >= 2 && parseInt(parts[0]) === currentYear && parseInt(parts[1]) === currentMonth) {
              income += inc.amount;
          }
      });
      
      expenses.forEach(exp => {
          const parts = exp.date.split('-');
          if (parts.length >= 2 && parseInt(parts[0]) === currentYear && parseInt(parts[1]) === currentMonth) {
              expense += exp.amount;
          }
      });
      
      return { income, expense, net: income - expense };
  })();

  $: budgetAlerts = budgetsData.filter(b => (b.spent / b.limit) >= 0.7);

  $: recentTransactions = (() => {
      const all = [
          ...incomes.map(i => ({ ...i, type: 'INCOME' })),
          ...expenses.map(e => ({ ...e, type: 'EXPENSE' }))
      ];
      // Sort by date desc
      all.sort((a, b) => new Date(b.date) - new Date(a.date));
      return all.slice(0, 5);
  })();

  // --- HELPERS ---
  let deleteConfirmTxId = null;
  function askDeleteTransaction(id) { deleteConfirmTxId = id; }
  function cancelDelete() { deleteConfirmTxId = null; }
  async function confirmDeleteTransaction(id) {
      await deleteTransaction(id);
      deleteConfirmTxId = null;
  }

  function formatDate(dateStr) {
      if (!dateStr) return '';
      const parts = dateStr.split('-');
      if (parts.length < 3) return dateStr;
      const day = parseInt(parts[2].split('T')[0]); // Handle possible timestamp
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des'];
      const month = months[parseInt(parts[1]) - 1] || '';
      const year = parts[0];
      return `${day} ${month} ${year}`;
  }

  const methodColors = {
      'CASH': 'green',
      'BANK': 'blue',
      'EWALLET': 'purple',
      'GOPAY': 'purple',
      'OVO': 'purple',
      'DANA': 'purple',
      'BCA': 'blue',
      'MANDIRI': 'blue',
      'CIMB': 'blue',
      'DEFAULT': 'default'
  };
  function getMethodColor(method) {
      if (!method) return 'default';
      const m = method.toUpperCase();
      return methodColors[m] || 'default';
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

  // Inline Editing Functions
  function enterEditMode(tx, field) {
      editingTxId = tx.id;
      editingField = field;
      editingValue = tx[field];
  }
  
  async function saveEdit(tx) {
      if (!editingTxId) return;
      
      const updatedTx = { ...tx };
      // Assign the new value
      if (editingField === 'amount') {
          updatedTx.amount = parseFloat(editingValue) || 0;
      } else {
          updatedTx[editingField] = editingValue;
      }

      try {
          const res = await fetch('/api/transactions/update', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(updatedTx)
          });
          if (res.ok) {
              await fetchAllData();
          }
      } catch (err) {}
      
      editingTxId = null;
      editingField = null;
  }
  
  function handleEditKeydown(e, tx) {
      if (e.key === 'Enter') {
          saveEdit(tx);
      } else if (e.key === 'Escape') {
          editingTxId = null;
          editingField = null;
      }
  }

  async function deleteTransaction(id) {
      try {
          const res = await fetch('/api/transactions/delete', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ id })
          });
          if (res.ok) {
              await fetchAllData();
          }
      } catch (err) {}
  }

  async function submitInlineTx(inlineTxObj, resetFunc) {
      if (!inlineTxObj.category || !inlineTxObj.amount || !inlineTxObj.method) return;
      const newTx = {
          id: "tx_" + Date.now(),
          date: inlineTxObj.date,
          type: inlineTxObj.type,
          category: inlineTxObj.category,
          amount: parseFloat(inlineTxObj.amount),
          description: inlineTxObj.description || "",
          method: inlineTxObj.method.toUpperCase()
      };
      try {
          const endpoint = newTx.type === 'INCOME' ? '/api/incomes/add' : '/api/expenses/add';
          const res = await fetch(endpoint, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify(newTx)
          });
          if (res.ok) {
              resetFunc();
              await fetchAllData();
          }
      } catch (err) {}
  }

  function submitInlineIncome() {
      submitInlineTx(newInlineIncome, () => {
          newInlineIncome.category = ''; newInlineIncome.amount = ''; newInlineIncome.method = ''; newInlineIncome.description = '';
      });
  }

  function submitInlineExpense() {
      submitInlineTx(newInlineExpense, () => {
          newInlineExpense.category = ''; newInlineExpense.amount = ''; newInlineExpense.method = ''; newInlineExpense.description = '';
      });
  }

  function handleInlineNewIncomeKeydown(e) {
      if (e.key === 'Enter') submitInlineIncome();
  }

  function handleInlineNewExpenseKeydown(e) {
      if (e.key === 'Enter') submitInlineExpense();
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
      } else if (tab === 'cashflow') {
          renderCashflowCharts();
      } else if (tab === 'portfolio') {
          renderPortfolioChart();
      }
  }

  function renderCashflowCharts() {
      if(!cfSavingsCanvas || !cfIncDonutCanvas || !cfExpDonutCanvas) return;

      if(cfSavingsChart) cfSavingsChart.destroy();
      cfSavingsChart = new Chart(cfSavingsCanvas, {
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
          options: { plugins: { legend: { labels: { color: 'white' } } }, scales: { x: { ticks: { color: '#94a3b8' } }, y: { ticks: { color: '#94a3b8' } } } }
      });

      let incMap = {};
      incomes.forEach(i => incMap[i.category] = (incMap[i.category] || 0) + i.amount);
      if(cfIncDonutChart) cfIncDonutChart.destroy();
      cfIncDonutChart = new Chart(cfIncDonutCanvas, {
          type: 'doughnut',
          data: { labels: Object.keys(incMap).length > 0 ? Object.keys(incMap) : ['No Data'], datasets: [{ data: Object.keys(incMap).length > 0 ? Object.values(incMap) : [1], backgroundColor: ['#34d399', '#10b981', '#059669'], borderWidth: 0 }] },
          options: { plugins: { legend: { display: false } }, cutout: '70%' }
      });

      let expMap = {};
      expenses.forEach(e => expMap[e.category] = (expMap[e.category] || 0) + e.amount);
      if(cfExpDonutChart) cfExpDonutChart.destroy();
      cfExpDonutChart = new Chart(cfExpDonutCanvas, {
          type: 'doughnut',
          data: { labels: Object.keys(expMap).length > 0 ? Object.keys(expMap) : ['No Data'], datasets: [{ data: Object.keys(expMap).length > 0 ? Object.values(expMap) : [1], backgroundColor: ['#f87171', '#ef4444', '#dc2626', '#b91c1c'], borderWidth: 0 }] },
          options: { plugins: { legend: { display: false } }, cutout: '70%' }
      });
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

<div class="app-layout">
  <!-- Sidebar -->
  <aside class="sidebar">
    <div class="sidebar-header">
      <h1 class="logo">💎 FinFlow</h1>
    </div>
    <div class="sidebar-menu">
      <button class:active={activeTab === 'dashboard'} on:click={() => setTab('dashboard')}>
          <span class="icon">📊</span> Dashboard
      </button>
      <button class:active={activeTab === 'cashflow'} on:click={() => setTab('cashflow')}>
          <span class="icon">💸</span> Cashflow
      </button>
      <button class:active={activeTab === 'portfolio'} on:click={() => setTab('portfolio')}>
          <span class="icon">💼</span> Portfolio
      </button>
    </div>
    <div class="sidebar-footer">
      <span class="version">v1.0</span>
    </div>
  </aside>

  <!-- Main Content Area -->
  <div class="main-area">
    <!-- Navbar -->
    <nav class="navbar glass-panel">
      <div class="nav-title">
        {activeTab === 'dashboard' ? 'Dashboard Overview' : activeTab === 'cashflow' ? 'Cashflow Management' : 'Portfolio Tracking'}
      </div>
      <div class="nav-stats">
        <span class="stat-badge"><span class="icon">💼</span> Kekayaan Bersih: Rp {netWorth.toLocaleString('id-ID')}</span>
      </div>
    </nav>

    <!-- Scrollable Content -->
    <main class="content-container fade-in">
      {#if activeTab === 'dashboard'}
         <div class="dashboard-grid">
             <!-- Row 1: Summary Cards -->
             <div class="summary-cards">
                 <div class="card glass-panel stat-card">
                     <h4>Pemasukan Bulan Ini</h4>
                     <div class="value text-green">Rp {currentMonthSummary.income.toLocaleString('id-ID')}</div>
                 </div>
                 <div class="card glass-panel stat-card">
                     <h4>Pengeluaran Bulan Ini</h4>
                     <div class="value text-red">Rp {currentMonthSummary.expense.toLocaleString('id-ID')}</div>
                 </div>
                 <div class="card glass-panel stat-card">
                     <h4>Sisa (Net) Bulan Ini</h4>
                     <div class="value" style="color: {currentMonthSummary.net >= 0 ? '#34d399' : '#f87171'}">Rp {currentMonthSummary.net.toLocaleString('id-ID')}</div>
                 </div>
             </div>

             <!-- Row 2: Alerts and Recent Transactions -->
             <div class="row" style="margin-top: 20px; align-items: flex-start;">
                 <div class="card glass-panel flex-1">
                     <h2>🚨 Budget Alerts</h2>
                     {#if budgetAlerts.length > 0}
                         <ul class="alert-list" style="margin-top:15px; padding-left:0; list-style:none;">
                             {#each budgetAlerts as alert}
                             <li style="margin-bottom: 12px; display:flex; justify-content:space-between; border-bottom:1px solid rgba(255,255,255,0.05); padding-bottom:8px;">
                                 <span class="category">{alert.category}</span>
                                 <span class="progress-text" style="color: { (alert.spent/alert.limit) > 0.9 ? '#ef4444' : '#facc15' }; font-weight:600;">{Math.round((alert.spent/alert.limit)*100)}% terpakai</span>
                             </li>
                             {/each}
                         </ul>
                     {:else}
                         <p class="text-muted" style="margin-top:10px;">Semua anggaran aman bulan ini.</p>
                     {/if}
                 </div>
                 
                 <div class="card glass-panel flex-2">
                     <h2>Recent Transactions</h2>
                     <table class="notion-table" style="margin-top:15px;">
                         <tbody>
                             {#each recentTransactions as tx}
                             <tr>
                                 <td style="width: 120px;">{formatDate(tx.date)}</td>
                                 <td>{tx.description} <span class="tag tag-{getMethodColor(tx.method)}" style="margin-left: 8px;">{tx.method}</span></td>
                                 <td style="width: 150px; text-align: right; color: {tx.type === 'INCOME' ? '#34d399' : '#f87171'}; font-weight: 500;">
                                     {tx.type === 'INCOME' ? '+' : '-'} Rp {tx.amount.toLocaleString('id-ID')}
                                 </td>
                             </tr>
                             {/each}
                         </tbody>
                     </table>
                 </div>
             </div>
         </div>

    {:else if activeTab === 'cashflow'}
       <div class="vertical-stack">
           <!-- 1. Chart Overview -->
           <div class="row-charts">
               <div class="card glass-panel chart-50">
                   <h2>Savings Trend (2026)</h2>
                   <canvas bind:this={cfSavingsCanvas}></canvas>
               </div>
               <div class="card glass-panel chart-25 center-content">
                   <h2>Income Breakdown</h2>
                   <div class="donut-wrapper"><canvas bind:this={cfIncDonutCanvas}></canvas></div>
               </div>
               <div class="card glass-panel chart-25 center-content">
                   <h2>Expense Breakdown</h2>
                   <div class="donut-wrapper"><canvas bind:this={cfExpDonutCanvas}></canvas></div>
               </div>
           </div>

           <!-- 2. Total Saving -->
           <div class="card glass-panel" style="margin-top: 20px;">
               <h2>Total Saving (Monthly)</h2>
               <div class="grid-12" style="margin-top: 15px;">
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
           </div>

           <!-- 3. Tabel Pemasukan -->
           <div class="card glass-panel" style="margin-top: 20px;">
               <h2>Pemasukan (Income)</h2>
               <div class="table-container notion-table-container">
                   <table class="notion-table">
                       <thead><tr><th>Tanggal</th><th>Keterangan (Source)</th><th>Kategori</th><th>Rekening</th><th>Nominal</th><th></th></tr></thead>
                       <tbody>
                           {#each incomes as tx}
                           <tr class="editable-row">
                               <td on:click={() => enterEditMode(tx, 'date')} style="cursor: pointer; width: 120px;">
                                   {#if editingTxId === tx.id && editingField === 'date'}
                                       <input type="date" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       {formatDate(tx.date)}
                                   {/if}
                               </td>
                               <td on:click={() => enterEditMode(tx, 'description')} style="cursor: pointer;">
                                   {#if editingTxId === tx.id && editingField === 'description'}
                                       <input type="text" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       {tx.description || '-'}
                                   {/if}
                               </td>
                               <td on:click={() => enterEditMode(tx, 'category')} style="cursor: pointer; width: 150px;">
                                   {#if editingTxId === tx.id && editingField === 'category'}
                                       <input type="text" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       <span class="tag tag-{getCategoryColor(tx.category)}">{tx.category}</span>
                                   {/if}
                               </td>
                               <td on:click={() => enterEditMode(tx, 'method')} style="cursor: pointer; width: 150px;">
                                   {#if editingTxId === tx.id && editingField === 'method'}
                                       <input type="text" list="methods-list" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       <span class="tag tag-{getMethodColor(tx.method)}">{tx.method}</span>
                                   {/if}
                               </td>
                               <td class="amount" style="cursor: pointer; color: #34d399; width: 150px;" on:click={() => enterEditMode(tx, 'amount')}>
                                   {#if editingTxId === tx.id && editingField === 'amount'}
                                       <input type="number" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       Rp {tx.amount.toLocaleString('id-ID')}
                                   {/if}
                               </td>
                               <td style="text-align: right; display: flex; gap: 4px; justify-content: flex-end; width: max-content;">
                                   {#if deleteConfirmTxId === tx.id}
                                       <button class="btn-danger btn-small" style="padding: 2px 6px; font-size: 11px;" on:click={() => confirmDeleteTransaction(tx.id)}>Ya</button>
                                       <button class="btn-secondary btn-small" style="padding: 2px 6px; font-size: 11px; background: rgba(255,255,255,0.1); color: white; border: none; border-radius: 4px; cursor: pointer;" on:click={cancelDelete}>Batal</button>
                                   {:else}
                                       <button class="btn-danger btn-small" on:click={() => askDeleteTransaction(tx.id)} title="Hapus">🗑️</button>
                                   {/if}
                               </td>
                           </tr>
                           {/each}
                           <!-- New Inline Row Income -->
                           <tr class="new-inline-row notion-new-row">
                               <td><input type="date" bind:value={newInlineIncome.date} /></td>
                               <td><input type="text" bind:value={newInlineIncome.description} placeholder="Keterangan..." /></td>
                               <td><input type="text" bind:value={newInlineIncome.category} placeholder="Kategori..." /></td>
                               <td>
                                   <input type="text" list="methods-list" bind:value={newInlineIncome.method} placeholder="Rekening..." />
                                   <datalist id="methods-list">
                                       {#each uniqueMethods as m} <option value={m}></option> {/each}
                                   </datalist>
                               </td>
                               <td><input type="number" bind:value={newInlineIncome.amount} on:keydown={handleInlineNewIncomeKeydown} placeholder="Nominal..." /></td>
                               <td style="text-align: right;">
                                   <button class="btn-primary btn-small" on:click={submitInlineIncome}>Add</button>
                               </td>
                           </tr>
                       </tbody>
                   </table>
               </div>
           </div>

           <!-- 4. Tabel Pengeluaran -->
           <div class="card glass-panel" style="margin-top: 20px;">
               <h2>Pengeluaran (Expense)</h2>
               <div class="table-container notion-table-container">
                   <table class="notion-table">
                       <thead><tr><th>Tanggal</th><th>Keterangan (Source)</th><th>Kategori</th><th>Rekening</th><th>Nominal</th><th></th></tr></thead>
                       <tbody>
                           {#each expenses as tx}
                           <tr class="editable-row">
                               <td on:click={() => enterEditMode(tx, 'date')} style="cursor: pointer; width: 120px;">
                                   {#if editingTxId === tx.id && editingField === 'date'}
                                       <input type="date" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       {formatDate(tx.date)}
                                   {/if}
                               </td>
                               <td on:click={() => enterEditMode(tx, 'description')} style="cursor: pointer;">
                                   {#if editingTxId === tx.id && editingField === 'description'}
                                       <input type="text" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       {tx.description || '-'}
                                   {/if}
                               </td>
                               <td on:click={() => enterEditMode(tx, 'category')} style="cursor: pointer; width: 150px;">
                                   {#if editingTxId === tx.id && editingField === 'category'}
                                       <input type="text" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       <span class="tag tag-{getCategoryColor(tx.category)}">{tx.category}</span>
                                   {/if}
                               </td>
                               <td on:click={() => enterEditMode(tx, 'method')} style="cursor: pointer; width: 150px;">
                                   {#if editingTxId === tx.id && editingField === 'method'}
                                       <input type="text" list="methods-list" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       <span class="tag tag-{getMethodColor(tx.method)}">{tx.method}</span>
                                   {/if}
                               </td>
                               <td class="amount" style="cursor: pointer; color: #f87171; width: 150px;" on:click={() => enterEditMode(tx, 'amount')}>
                                   {#if editingTxId === tx.id && editingField === 'amount'}
                                       <input type="number" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       Rp {tx.amount.toLocaleString('id-ID')}
                                   {/if}
                               </td>
                               <td style="text-align: right; display: flex; gap: 4px; justify-content: flex-end; width: max-content;">
                                   {#if deleteConfirmTxId === tx.id}
                                       <button class="btn-danger btn-small" style="padding: 2px 6px; font-size: 11px;" on:click={() => confirmDeleteTransaction(tx.id)}>Ya</button>
                                       <button class="btn-secondary btn-small" style="padding: 2px 6px; font-size: 11px; background: rgba(255,255,255,0.1); color: white; border: none; border-radius: 4px; cursor: pointer;" on:click={cancelDelete}>Batal</button>
                                   {:else}
                                       <button class="btn-danger btn-small" on:click={() => askDeleteTransaction(tx.id)} title="Hapus">🗑️</button>
                                   {/if}
                               </td>
                           </tr>
                           {/each}
                           <!-- New Inline Row Expense -->
                           <tr class="new-inline-row notion-new-row">
                               <td><input type="date" bind:value={newInlineExpense.date} /></td>
                               <td><input type="text" bind:value={newInlineExpense.description} placeholder="Keterangan..." /></td>
                               <td><input type="text" bind:value={newInlineExpense.category} placeholder="Kategori..." /></td>
                               <td><input type="text" list="methods-list" bind:value={newInlineExpense.method} placeholder="Rekening..." /></td>
                               <td><input type="number" bind:value={newInlineExpense.amount} on:keydown={handleInlineNewExpenseKeydown} placeholder="Nominal..." /></td>
                               <td style="text-align: right;">
                                   <button class="btn-primary btn-small" on:click={submitInlineExpense}>Add</button>
                               </td>
                           </tr>
                       </tbody>
                   </table>
               </div>
           </div>

           <!-- 5. Tabel Budget -->
           <div class="card glass-panel" style="margin-top: 20px;">
               <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">
                   <h2>Anggaran & Realisasi (Budget)</h2>
               </div>
               <div class="table-container notion-table-container">
                   <table class="notion-table">
                       <thead><tr><th>Kategori</th><th>Interval</th><th>Limit Anggaran</th><th>Terpakai</th><th>Sisa</th><th>Realisasi</th></tr></thead>
                       <tbody>
                           {#each budgetsData as b}
                           <tr>
                               <td><span class="tag tag-{getCategoryColor(b.category)}">{b.category}</span></td>
                               <td><span class="tag tag-default">{b.interval}</span></td>
                               <td class="amount">Rp {b.limit.toLocaleString('id-ID')}</td>
                               <td class="amount">Rp {b.spent.toLocaleString('id-ID')}</td>
                               <td class="amount">Rp {b.remaining.toLocaleString('id-ID')}</td>
                               <td style="width: 150px;">
                                   <div style="display: flex; align-items: center; gap: 8px;">
                                       <div class="budget-bar-bg" style="width: 100%; height: 8px; background: rgba(255,255,255,0.1); border-radius: 4px; overflow: hidden; flex: 1;">
                                           <div class="budget-bar-fill" style="width: {Math.min((b.spent/b.limit)*100, 100)}%; height: 100%; background: {(b.spent/b.limit) > 0.9 ? '#ef4444' : (b.spent/b.limit) > 0.7 ? '#facc15' : '#10b981'};"></div>
                                       </div>
                                       <span style="font-size: 0.8rem; color: #94a3b8; width: 35px; text-align: right;">{Math.round((b.spent/b.limit)*100)}%</span>
                                   </div>
                               </td>
                           </tr>
                           {/each}
                           <!-- Set budget inline form -->
                           <tr class="new-inline-row notion-new-row">
                               <td><input type="text" bind:value={newBudgetCategory} placeholder="Kategori Baru..." /></td>
                               <td>
                                   <select bind:value={newBudgetInterval}>
                                       <option value="MONTHLY">Bulanan</option>
                                       <option value="WEEKLY">Mingguan</option>
                                   </select>
                               </td>
                               <td><input type="number" bind:value={newBudgetLimit} placeholder="Limit (Rp)" /></td>
                               <td colspan="3" style="text-align: right;">
                                   <button class="btn-primary btn-small" on:click={submitBudget}>Set Budget</button>
                               </td>
                           </tr>
                       </tbody>
                   </table>
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
    </main>
  </div>
</div>

<style>
  :global(body) { background-color: #0f172a; color: #f8fafc; font-family: 'Inter', sans-serif; margin: 0; padding: 0; height: 100vh; overflow: hidden; }
  h1, h2, h3, h4 { margin-top: 0; }
  h2 { font-size: 1.1rem; color: #e2e8f0; font-weight: 600; }
  
  /* Layout */
  .app-layout { display: flex; height: 100vh; width: 100vw; overflow: hidden; }
  .sidebar { width: 250px; background: rgba(15, 23, 42, 0.95); border-right: 1px solid rgba(255,255,255,0.05); display: flex; flex-direction: column; padding: 20px; z-index: 10; box-sizing: border-box; }
  .sidebar-header { margin-bottom: 30px; }
  .sidebar-header .logo { font-size: 1.5rem; text-align: left; margin: 0; background: linear-gradient(to right, #38bdf8, #818cf8); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
  
  .sidebar-menu { display: flex; flex-direction: column; gap: 8px; flex: 1; }
  .sidebar-menu button { display: flex; align-items: center; gap: 12px; background: transparent; border: none; padding: 12px 16px; border-radius: 8px; color: #94a3b8; font-weight: 500; cursor: pointer; transition: 0.2s; font-size: 0.95rem; text-align: left; }
  .sidebar-menu button:hover { background: rgba(255,255,255,0.03); color: #e2e8f0; }
  .sidebar-menu button.active { background: rgba(56, 189, 248, 0.1); color: #38bdf8; font-weight: 600; }
  .sidebar-menu button .icon { font-size: 1.1rem; }
  
  .sidebar-footer { font-size: 0.8rem; color: #64748b; }
  
  .main-area { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: #0f172a; }
  
  .navbar { display: flex; justify-content: space-between; align-items: center; padding: 15px 30px; border-bottom: 1px solid rgba(255,255,255,0.05); border-radius: 0; margin-bottom: 0; box-shadow: none; z-index: 5; background: rgba(15, 23, 42, 0.8); backdrop-filter: blur(10px); }
  .navbar .nav-title { font-size: 1.2rem; font-weight: 600; color: #e2e8f0; }
  .navbar .nav-stats .stat-badge { background: rgba(56, 189, 248, 0.1); border: 1px solid rgba(56, 189, 248, 0.2); padding: 8px 16px; border-radius: 20px; font-weight: 600; color: #38bdf8; font-size: 0.9rem; }
  
  .content-container { flex: 1; overflow-y: auto; padding: 30px; }
  
  .dashboard-grid { display: flex; flex-direction: column; gap: 20px; }
  .summary-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
  .stat-card h4 { margin: 0 0 10px 0; color: #94a3b8; font-size: 0.9rem; font-weight: 500; }
  .stat-card .value { font-size: 1.5rem; font-weight: 700; }

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
  
  /* Vertical Stack Layout */
  .vertical-stack { display: flex; flex-direction: column; gap: 20px; }
  .row-charts { display: flex; gap: 20px; width: 100%; }
  .chart-50 { flex: 2; }
  .chart-25 { flex: 1; }

  /* Tables */
  .minimal-table { width: 100%; border-collapse: collapse; }
  .minimal-table td { padding: 8px 0; border-bottom: 1px solid rgba(255,255,255,0.05); }
  .minimal-table .amount { text-align: right; font-weight: bold; }
  
  .full-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  .full-table th, .full-table td { padding: 12px 10px; border-bottom: 1px solid rgba(255,255,255,0.05); text-align: left; }
  .full-table th { color: #94a3b8; font-weight: 500; }
  .full-table .amount { text-align: right; font-family: monospace; font-size: 1rem; }
  .table-container { max-height: 400px; overflow-y: auto; }

  /* Notion Style Table */
  .notion-table-container { max-height: 500px; overflow-y: auto; border: 1px solid rgba(255, 255, 255, 0.08); border-radius: 6px; }
  .notion-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
  .notion-table th { background: rgba(255,255,255,0.02); color: #94a3b8; font-weight: 500; text-align: left; padding: 8px 12px; border-bottom: 1px solid rgba(255, 255, 255, 0.08); border-right: 1px solid rgba(255, 255, 255, 0.05); }
  .notion-table td { padding: 6px 12px; border-bottom: 1px solid rgba(255, 255, 255, 0.05); border-right: 1px solid rgba(255, 255, 255, 0.02); }
  .notion-table td:last-child, .notion-table th:last-child { border-right: none; }
  .notion-table .amount { font-family: monospace; font-size: 0.95rem; }
  .notion-table .editable-row:hover { background: rgba(255,255,255,0.02); }
  .notion-table .editable-row input { border: none; background: transparent; padding: 0; margin: 0; color: inherit; font-size: inherit; font-family: inherit; width: 100%; box-sizing: border-box; }
  .notion-table .editable-row input:focus { outline: none; border-bottom: 1px solid #38bdf8; }
  .notion-new-row td { background: rgba(255,255,255,0.01); }
  .notion-new-row input, .notion-new-row select { padding: 4px 8px; font-size: 0.85rem; background: transparent; border: 1px solid rgba(255,255,255,0.1); width: 100%; box-sizing: border-box; }
  
  /* Filter */
  .filter-group { display: flex; gap: 5px; }
  .filter-group button { background: none; border: 1px solid rgba(255,255,255,0.2); color: #94a3b8; border-radius: 4px; padding: 4px 8px; cursor: pointer; font-size: 0.8rem; }
  .filter-group button.active { background: #38bdf8; color: #0f172a; border-color: #38bdf8; font-weight: bold; }
  
  /* Tag Capsules */
  .tag { display: inline-block; padding: 3px 8px; border-radius: 12px; font-size: 0.75rem; font-weight: 500; white-space: nowrap; }
  .tag-default { background: rgba(148, 163, 184, 0.2); color: #cbd5e1; }
  .tag-blue { background: rgba(56, 189, 248, 0.2); color: #7dd3fc; }
  .tag-green { background: rgba(52, 211, 153, 0.2); color: #6ee7b7; }
  .tag-yellow { background: rgba(250, 204, 21, 0.2); color: #fde047; }
  .tag-red { background: rgba(248, 113, 113, 0.2); color: #fca5a5; }
  .tag-purple { background: rgba(192, 132, 252, 0.2); color: #d8b4fe; }
  .tag-pink { background: rgba(244, 114, 182, 0.2); color: #f9a8d4; }
  .tag-indigo { background: rgba(129, 140, 248, 0.2); color: #a5b4fc; }
</style>
