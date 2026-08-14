<script>
  import { onMount, tick } from 'svelte';
  import Chart from 'chart.js/auto';
  import { 
      LayoutDashboard, 
      ArrowLeftRight, 
      Briefcase, 
      Palette, 
      Sun, 
      Moon, 
      TrendingUp, 
      TrendingDown, 
      AlertTriangle, 
      Trash2, 
      Plus, 
      Edit2,
      Check,
      X,
      Lightbulb,
      FileText
  } from 'lucide-svelte';

  // --- STATE ---
  let activeTab = 'dashboard';
  let currentTheme = 'dark';
  let expenses = [];
  let incomes = [];
  let portfolioData = null;
  let accountsData = [];
  let budgetsData = [];
  
  // Charts
  let portfolioCanvas;
  let portfolioChart = null;

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
  $: totalFloatingPnL = portfolioData?.assets ? portfolioData.assets.reduce((sum, a) => sum + a.profit_loss, 0) : 0;
  
  let monthlyData = Array.from({length: 12}, (_, i) => ({ month: i+1, income: 0, expense: 0, net: 0 }));

  // --- LIFECYCLE ---
  onMount(async () => {
    currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
    await fetchAllData();
  });

  function toggleTheme() {
    currentTheme = currentTheme === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', currentTheme);
    try { localStorage.setItem('theme', currentTheme); } catch (e) { console.warn('Unable to persist theme preference', e); }
  }

  // --- FETCHING ---
  async function fetchAllData() {
    await Promise.all([
        fetchExpenses(),
        fetchIncomes(),
        fetchPortfolio(),
        fetchAccounts(),
        fetchBudgets(),
        fetchAnalytics()
    ]);
    calculateMonthlyData();
    if (activeTab === 'cashflow') {
        setTimeout(renderCashflowCharts, 100);
    } else if (activeTab === 'portfolio') {
        setTimeout(renderPortfolioChart, 100);
    }
  }

  async function fetchExpenses() { try { const res = await fetch('/api/expenses'); if (res.ok) expenses = await res.json() || []; } catch(e) { console.error('fetchExpenses:', e); } }
  async function fetchIncomes() { try { const res = await fetch('/api/incomes'); if (res.ok) incomes = await res.json() || []; } catch(e) { console.error('fetchIncomes:', e); } }
  async function fetchPortfolio() { try { const res = await fetch('/api/portfolio'); if (res.ok) portfolioData = await res.json(); } catch(e) { console.error('fetchPortfolio:', e); } }
  async function fetchAccounts() { try { const res = await fetch('/api/accounts'); if (res.ok) accountsData = await res.json() || []; } catch(e) { console.error('fetchAccounts:', e); } }
  async function fetchBudgets() { try { const res = await fetch('/api/budgets'); if (res.ok) budgetsData = await res.json() || []; } catch(e) { console.error('fetchBudgets:', e); } }
  async function fetchAnalytics() { try { const res = await fetch('/api/analytics'); if (res.ok) analyticsData = await res.json(); } catch(e) { console.error('fetchAnalytics:', e); } }

  function calculateMonthlyData() {
      const currentYear = new Date().getFullYear();
      // Reset
      monthlyData = Array.from({length: 12}, (_, i) => ({ month: i+1, income: 0, expense: 0, net: 0 }));
      
      incomes.forEach(inc => {
          const parts = inc.date.split('-');
          const y = parseInt(parts[0]);
          const m = parseInt(parts[1]);
          if(y === currentYear && m >= 1 && m <= 12) monthlyData[m-1].income += inc.amount;
      });
      expenses.forEach(exp => {
          const parts = exp.date.split('-');
          const y = parseInt(parts[0]);
          const m = parseInt(parts[1]);
          if(y === currentYear && m >= 1 && m <= 12) monthlyData[m-1].expense += exp.amount;
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



  // --- CHARTS ---
  async function setTab(tab) {
      activeTab = tab;
      await tick();
      if (tab === 'cashflow') {
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


  
  // Helpers

</script>

<div class="app-layout">
  <!-- Sidebar -->
  <aside class="sidebar">
    <div class="sidebar-header">
      <h1 class="logo" style="display: flex; align-items: center; gap: 8px;"><LayoutDashboard size={24} color="var(--brand-primary)" /> FinFlow</h1>
    </div>
    <div class="sidebar-menu">
      <button class:active={activeTab === 'dashboard'} on:click={() => setTab('dashboard')}>
          <span class="icon" style="display: flex;"><LayoutDashboard size={18} /></span> Dashboard
      </button>
      <button class:active={activeTab === 'cashflow'} on:click={() => setTab('cashflow')}>
          <span class="icon" style="display: flex;"><ArrowLeftRight size={18} /></span> Cashflow
      </button>
      <button class:active={activeTab === 'portfolio'} on:click={() => setTab('portfolio')}>
          <span class="icon" style="display: flex;"><Briefcase size={18} /></span> Portfolio
      </button>
      <button class:active={activeTab === 'design-system'} on:click={() => setTab('design-system')}>
          <span class="icon" style="display: flex;"><Palette size={18} /></span> Design System
      </button>
    </div>
    <div class="sidebar-footer">
      <button class="theme-btn" on:click={toggleTheme} style="background:transparent; border:1px solid var(--border-color); color:var(--text-secondary); width:100%; text-align:left; padding:8px 12px; border-radius:var(--radius-sm); margin-bottom:12px; cursor:pointer; display:flex; align-items:center; gap:8px;">
          {#if currentTheme === 'dark'}
              <span class="icon" style="display: flex;"><Sun size={16} /></span> Light Mode
          {:else}
              <span class="icon" style="display: flex;"><Moon size={16} /></span> Dark Mode
          {/if}
      </button>
      <span class="version" style="display:block; text-align:center;">v1.0</span>
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
        <span class="stat-badge" style="display: inline-flex; align-items: center; gap: 6px;"><Briefcase size={16} /> Kekayaan Bersih: Rp {netWorth.toLocaleString('id-ID')}</span>
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
                     <div class="value" style="color: {currentMonthSummary.net >= 0 ? 'var(--color-green)' : 'var(--color-red)'}">Rp {currentMonthSummary.net.toLocaleString('id-ID')}</div>
                 </div>
             </div>

             <!-- Financial Health Analytics -->
             <div class="row" style="margin-top: 20px;">
                 <div class="card glass-panel flex-1">
                     <h2 style="display: flex; align-items: center; gap: 8px;"><LayoutDashboard size={20} /> Financial Health</h2>
                     <div style="margin-top: 15px; display: flex; flex-direction: column; gap: 10px;">
                         <div>
                             <div class="flex-between">
                                 <span>Savings Rate</span>
                                 <span style="font-weight: bold; color: {analyticsData.savings_rate >= 20 ? 'var(--color-green)' : 'var(--color-yellow)'};">
                                     {analyticsData.savings_rate ? analyticsData.savings_rate.toFixed(1) : 0}%
                                 </span>
                             </div>
                             <div style="font-size: 0.8rem; color: var(--text-secondary);">Target: > 20%</div>
                         </div>
                         <div>
                             <div class="flex-between">
                                 <span>Emergency Fund</span>
                                 <span style="font-weight: bold; color: {analyticsData.emergency_run_rate >= 6 ? 'var(--color-green)' : 'var(--color-red)'};">
                                     {analyticsData.emergency_run_rate ? analyticsData.emergency_run_rate.toFixed(1) : 0} bulan
                                 </span>
                             </div>
                             <div style="font-size: 0.8rem; color: var(--text-secondary);">Target: > 6 bulan pengeluaran</div>
                         </div>
                         <div style="margin-top: 10px; padding: 10px; background: var(--bg-secondary); border-radius: var(--radius-sm); border-left: 3px solid var(--brand-primary); font-size: 0.9rem; display: flex; align-items: flex-start; gap: 8px;">
                             <Lightbulb size={18} style="flex-shrink: 0; margin-top: 2px;" /> 
                             <span>{analyticsData.recommendation || 'Data belum tersedia'}</span>
                         </div>
                     </div>
                 </div>
             </div>

             <!-- Row 2: Alerts and Recent Transactions -->
             <div class="row" style="margin-top: 20px; align-items: flex-start;">
                 <div class="card glass-panel flex-1">
                     <h2 style="display: flex; align-items: center; gap: 8px;"><AlertTriangle size={20} /> Budget Alerts</h2>
                     {#if budgetAlerts.length > 0}
                         <ul class="alert-list" style="margin-top:15px; padding-left:0; list-style:none;">
                             {#each budgetAlerts as alert}
                             <li style="margin-bottom: 12px; display:flex; justify-content:space-between; border-bottom:1px solid var(--border-color); padding-bottom:8px;">
                                 <span class="category">{alert.category}</span>
                                 <span class="progress-text" style="color: { (alert.spent/alert.limit) > 0.9 ? 'var(--color-red)' : 'var(--color-yellow)' }; font-weight:600;">{Math.round((alert.spent/alert.limit)*100)}% terpakai</span>
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
                                 <td style="width: 150px; text-align: right; color: {tx.type === 'INCOME' ? 'var(--color-green)' : 'var(--color-red)'}; font-weight: 500;">
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
                   <h2 style="display: flex; align-items: center; gap: 8px;"><TrendingUp size={20} /> Savings Trend ({new Date().getFullYear()})</h2>
                   <canvas bind:this={cfSavingsCanvas}></canvas>
               </div>
               <div class="card glass-panel chart-25 center-content">
                   <h2 style="display: flex; align-items: center; gap: 8px;"><TrendingUp size={20} /> Income Breakdown</h2>
                   <div class="donut-wrapper"><canvas bind:this={cfIncDonutCanvas}></canvas></div>
               </div>
               <div class="card glass-panel chart-25 center-content">
                   <h2 style="display: flex; align-items: center; gap: 8px;"><TrendingDown size={20} /> Expense Breakdown</h2>
                   <div class="donut-wrapper"><canvas bind:this={cfExpDonutCanvas}></canvas></div>
               </div>
           </div>

           <!-- 2. Total Saving -->
           <div class="card glass-panel" style="margin-top: 20px;">
               <h2 style="display: flex; align-items: center; gap: 8px;"><ArrowLeftRight size={20} /> Total Saving (Monthly)</h2>
               <div class="grid-12" style="margin-top: 15px;">
                   {#each monthlyData as data}
                   <div class="month-card glass-panel">
                       <h4>Bulan {data.month}</h4>
                       <div class="stat"><span class="text-green">In:</span> Rp {data.income.toLocaleString('id-ID')}</div>
                       <div class="stat"><span class="text-red">Out:</span> Rp {data.expense.toLocaleString('id-ID')}</div>
                       <div class="stat-net" style="color: {data.net >= 0 ? 'var(--color-green)' : 'var(--color-red)'}">
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
                               <td class="amount" style="cursor: pointer; color: var(--color-green); width: 150px;" on:click={() => enterEditMode(tx, 'amount')}>
                                   {#if editingTxId === tx.id && editingField === 'amount'}
                                       <input type="number" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       Rp {tx.amount.toLocaleString('id-ID')}
                                   {/if}
                               </td>
                               <td style="text-align: right; display: flex; gap: 4px; justify-content: flex-end; width: max-content;">
                                   {#if deleteConfirmTxId === tx.id}
                                       <button class="btn-danger btn-small" style="padding: 2px 6px; font-size: 11px;" on:click={() => confirmDeleteTransaction(tx.id)}>Ya</button>
                                       <button class="btn-secondary btn-small" style="padding: 2px 6px; font-size: 11px; background: var(--border-color); color: var(--text-primary); border: none; border-radius: var(--radius-sm); cursor: pointer;" on:click={cancelDelete}>Batal</button>
                                   {:else}
                                        <button class="btn-danger btn-small" style="display: flex; align-items: center; justify-content: center; padding: 4px;" on:click={() => askDeleteTransaction(tx.id)} title="Hapus" aria-label="Hapus transaksi"><Trash2 size={14} /></button>
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
                               <td class="amount" style="cursor: pointer; color: var(--color-red); width: 150px;" on:click={() => enterEditMode(tx, 'amount')}>
                                   {#if editingTxId === tx.id && editingField === 'amount'}
                                       <input type="number" bind:value={editingValue} on:blur={() => saveEdit(tx)} on:keydown={(e) => handleEditKeydown(e, tx)} autofocus />
                                   {:else}
                                       Rp {tx.amount.toLocaleString('id-ID')}
                                   {/if}
                               </td>
                               <td style="text-align: right; display: flex; gap: 4px; justify-content: flex-end; width: max-content;">
                                   {#if deleteConfirmTxId === tx.id}
                                       <button class="btn-danger btn-small" style="padding: 2px 6px; font-size: 11px;" on:click={() => confirmDeleteTransaction(tx.id)}>Ya</button>
                                       <button class="btn-secondary btn-small" style="padding: 2px 6px; font-size: 11px; background: var(--border-color); color: var(--text-primary); border: none; border-radius: var(--radius-sm); cursor: pointer;" on:click={cancelDelete}>Batal</button>
                                   {:else}
                                       <button class="btn-danger btn-small" style="display: flex; align-items: center; justify-content: center; padding: 4px;" on:click={() => askDeleteTransaction(tx.id)} title="Hapus"><Trash2 size={14} /></button>
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
                                       <div class="budget-bar-bg" style="width: 100%; height: 8px; background: var(--border-color); border-radius: var(--radius-pill); overflow: hidden; flex: 1;">
                                           <div class="budget-bar-fill" style="width: {Math.min((b.spent/b.limit)*100, 100)}%; height: 100%; background: {(b.spent/b.limit) > 0.9 ? 'var(--color-red)' : (b.spent/b.limit) > 0.7 ? 'var(--color-yellow)' : 'var(--color-green)'};"></div>
                                       </div>
                                       <span style="font-size: 0.8rem; color: var(--text-secondary); width: 35px; text-align: right;">{Math.round((b.spent/b.limit)*100)}%</span>
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
                                <div style="display: flex; align-items: center; gap: 6px;">
                                    {asset.code}
                                    {#if asset.profit_loss_pct <= -10} <span title="Alert!" style="color: var(--color-yellow); display: flex;"><AlertTriangle size={14} /></span> {/if}
                                </div>
                             </td>
                             <td>
                                <div style="display: flex; align-items: center; gap: 6px;">
                                    {asset.type}
                                    {#if asset.type === 'REKSA_DANA'}
                                       <button class="btn-small" style="padding: 2px 5px; display: flex; align-items: center; justify-content: center;" on:click={() => updateNAV(asset.id, asset.current_price)}><Edit2 size={12} /></button>
                                    {/if}
                                </div>
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
    {:else if activeTab === 'design-system'}
       <div class="content-container fade-in">
           <h2 style="display: flex; align-items: center; gap: 8px;"><Palette size={24} /> Design System Preview</h2>
           <p class="text-secondary" style="margin-bottom:20px;">Pratinjau komponen-komponen UI yang diatur menggunakan Design System terpusat.</p>

           <div class="vertical-stack">
               <div class="glass-panel">
                   <h3 style="display: flex; align-items: center; gap: 8px;"><Palette size={20} /> Color Palette & Status Tags</h3>
                   <div style="display:flex; gap:10px; flex-wrap:wrap; margin-top:15px;">
                       <span class="tag tag-default">Default</span>
                       <span class="tag tag-blue">Primary Blue</span>
                       <span class="tag tag-green">Success Green</span>
                       <span class="tag tag-yellow">Warning Yellow</span>
                       <span class="tag tag-red">Danger Red</span>
                       <span class="tag tag-purple">Purple</span>
                   </div>
               </div>

               <div class="glass-panel">
                   <h3 style="display: flex; align-items: center; gap: 8px;"><FileText size={20} /> Typography</h3>
                   <h1>Heading 1 (h1)</h1>
                   <h2>Heading 2 (h2)</h2>
                   <h3>Heading 3 (h3)</h3>
                   <h4>Heading 4 (h4) - Subtitle</h4>
                   <p>This is standard body text. It uses the Inter font. Monospace text uses <span class="text-mono">JetBrains Mono</span>.</p>
                   <p class="text-secondary">This is secondary text, typically used for descriptions.</p>
                   <p class="text-muted">This is muted text, used for disabled or background info.</p>
               </div>

               <div class="glass-panel">
                   <h3 style="display: flex; align-items: center; gap: 8px;"><Plus size={20} /> Interactive Components</h3>
                   <div style="display:flex; gap:15px; margin-top:15px; align-items:center;">
                       <button class="btn-primary">Primary Button</button>
                       <button class="btn-primary" disabled>Disabled</button>
                       <button class="btn-small">Small Outline</button>
                   </div>
                   <div style="margin-top: 20px; max-width: 300px;">
                       <label style="display:block; margin-bottom:5px; font-size:0.9rem;" class="text-secondary">Example Input</label>
                       <input type="text" placeholder="Type something..." />
                   </div>
               </div>
               
               <div class="glass-panel" style="background: var(--brand-light); border-color: var(--brand-active);">
                   <h3 class="text-blue" style="display: flex; align-items: center; gap: 8px;"><Sun size={20} /> Glassmorphism Showcase</h3>
                   <p>This panel uses a tinted glassmorphism effect using CSS backdrop-filter.</p>
               </div>
           </div>
       </div>
    {/if}
    </main>
  </div>
</div>

<style>
  /* Layout Scope */
  .sidebar { width: 250px; background: var(--bg-sidebar); border-right: 1px solid var(--border-color); display: flex; flex-direction: column; padding: var(--spacing-lg); z-index: 10; box-sizing: border-box; }
  .sidebar-header { margin-bottom: 30px; }
  .sidebar-header .logo { font-size: 1.5rem; text-align: left; margin: 0; background: linear-gradient(to right, var(--brand-primary), var(--color-purple)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
  
  .sidebar-menu { display: flex; flex-direction: column; gap: 8px; flex: 1; }
  .sidebar-menu button { display: flex; align-items: center; gap: 12px; background: transparent; border: none; padding: 12px 16px; border-radius: var(--radius-sm); color: var(--text-secondary); font-weight: 500; cursor: pointer; transition: background var(--transition-fast), color var(--transition-fast); font-size: 0.95rem; text-align: left; }
  .sidebar-menu button:hover { background: var(--border-color); color: var(--text-primary); }
  .sidebar-menu button.active { background: var(--brand-light); color: var(--brand-primary); font-weight: 600; }
  .sidebar-menu button .icon { font-size: 1.1rem; }
  
  .sidebar-footer { font-size: 0.8rem; color: var(--text-muted); }
  
  .main-area { flex: 1; display: flex; flex-direction: column; overflow: hidden; background: var(--bg-primary); }
  
  .navbar { display: flex; justify-content: space-between; align-items: center; padding: 15px 30px; border-bottom: 1px solid var(--border-color); border-radius: 0; margin-bottom: 0; box-shadow: none; z-index: 5; background: var(--glass-bg); backdrop-filter: blur(12px); -webkit-backdrop-filter: blur(12px); }
  .navbar .nav-title { font-size: 1.2rem; font-weight: 600; color: var(--text-primary); }
  .navbar .nav-stats .stat-badge { background: var(--brand-light); border: 1px solid var(--brand-active); padding: 8px 16px; border-radius: var(--radius-pill); font-weight: 600; color: var(--brand-primary); font-size: 0.9rem; }
  
  .content-container { flex: 1; overflow-y: auto; padding: 30px; }
  
  .dashboard-grid { display: flex; flex-direction: column; gap: 20px; }
  .summary-cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; }
  .stat-card .value { font-size: 1.5rem; font-weight: 700; color: var(--text-primary); }

  .container { max-width: 1200px; margin: 0 auto; }
  .row { display: flex; gap: 20px; flex-wrap: wrap; }
  .col { display: flex; flex-direction: column; gap: 20px; }
  .flex-1 { flex: 1; } .flex-2 { flex: 2; }
  .flex-between { display: flex; justify-content: space-between; align-items: center; }
  
  /* Utilities */
  .center-content { display: flex; flex-direction: column; align-items: center; justify-content: center; }
  .fade-in { animation: fadeIn 0.3s ease-in; }
  @keyframes fadeIn { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
  
  /* Charts */
  .donut-wrapper { width: 160px; height: 160px; margin-top: 10px; }
  .chart-container { max-width: 250px; margin: 0 auto; }
  
  /* Grid 12 Months */
  .grid-12 { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; }
  .month-card { padding: 15px; text-align: center; }
  .month-card h4 { margin: 0 0 10px 0; color: var(--text-secondary); font-size: 0.9rem; }
  .month-card .stat { font-size: 0.85rem; margin-bottom: 5px; }
  .month-card .stat-net { margin-top: 10px; font-weight: bold; border-top: 1px solid var(--border-color); padding-top: 5px; }
  
  .form-row { display: flex; gap: 10px; margin-top: 15px; align-items: center; }
  
  /* Vertical Stack Layout */
  .vertical-stack { display: flex; flex-direction: column; gap: 20px; }
  .row-charts { display: flex; gap: 20px; width: 100%; }
  .chart-50 { flex: 2; }
  .chart-25 { flex: 1; }

  /* Tables */
  .full-table { width: 100%; border-collapse: collapse; font-size: 0.9rem; }
  .full-table th, .full-table td { padding: 12px 10px; border-bottom: 1px solid var(--border-color); text-align: left; }
  .full-table th { color: var(--text-secondary); font-weight: 500; }
  .table-container { max-height: 400px; overflow-y: auto; }

  /* Notion Style Table */
  .notion-table-container { max-height: 500px; overflow-y: auto; border: 1px solid var(--border-color); border-radius: var(--radius-sm); }
  .notion-table { width: 100%; border-collapse: collapse; font-size: 0.875rem; }
  .notion-table th { background: var(--bg-secondary); color: var(--text-secondary); font-weight: 500; text-align: left; padding: 8px 12px; border-bottom: 1px solid var(--border-color); border-right: 1px solid var(--border-color); }
  .notion-table td { padding: 6px 12px; border-bottom: 1px solid var(--border-color); border-right: 1px solid var(--border-color); }
  .notion-table td:last-child, .notion-table th:last-child { border-right: none; }
  .notion-table .amount { font-family: var(--font-mono); font-size: 0.95rem; }
  .notion-table .editable-row:hover { background: var(--bg-secondary); }
  .notion-table .editable-row input { border: none; background: transparent; padding: 0; margin: 0; color: inherit; font-size: inherit; font-family: inherit; width: 100%; box-sizing: border-box; }
  .notion-table .editable-row input:focus { outline: none; border-bottom: 1px solid var(--brand-primary); }
  .notion-new-row td { background: var(--bg-secondary); }
  .notion-new-row input, .notion-new-row select { padding: 4px 8px; font-size: 0.85rem; background: transparent; border: 1px solid var(--border-color); width: 100%; box-sizing: border-box; }
</style>
