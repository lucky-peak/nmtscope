import { useEffect, useState, useMemo } from "react";
import type { NMTReport } from "../types/nmt";
import {
    Legend,
    Line,
    LineChart,
    ResponsiveContainer,
    Tooltip,
    XAxis,
    YAxis
} from 'recharts';

// 定义颜色盘
const PALETTE = [
    "#1f77b4", "#ff7f0e", "#2ca02c", "#d62728", "#9467bd",
    "#8c564b", "#e377c2", "#7f7f7f", "#bcbd22", "#17becf",
    "#393b79", "#637939", "#8c6d31", "#843c39", "#7b4173"
];

// 定义排序的时间间隔 (秒)
const SORT_INTERVALS = {
    '10s': 10,
    '30s': 30,
    '1m': 60,
    '5m': 300,
    '15m': 900,
    '30m': 1800,
    '60m': 3600,
};

// ⭐ 更新: 定义快速选择的时间范围 (秒)，只保留 10m, 30m, 1h
const QUICK_RANGES = {
    '10m': 60 * 10,  // 600 seconds
    '30m': 60 * 30,  // 1800 seconds
    '1h': 3600,      // 3600 seconds
};

type SortCriteria = 'none' | keyof typeof SORT_INTERVALS;
type SortType = 'delta' | 'rate'; // 新增排序类型

// 用于存储排序列表项的类型，新增 rate 字段
type CategoryListItem = {
    name: string;
    delta: number | null; // Committed 绝对增量 (MB)
    rate: number | null;  // Committed 变化率 (MB/s)
};

// 检查数据是否足够进行特定排序的帮助函数
function isDataSufficient(nmt: NMTReport[], intervalSeconds: number): boolean {
    if (nmt.length < 2) return false;

    const latestTime = nmt[nmt.length - 1].created;
    const targetTime = latestTime - intervalSeconds;
    
    // 查找最接近目标时间点的数据
    let pastReport = nmt.reduce((prev, curr) => 
        (Math.abs(curr.created - targetTime) < Math.abs(prev.created - targetTime) ? curr : prev), 
        nmt[0]
    );

    const timeDeltaTolerance = 5; // 允许 5 秒的误差
    const actualInterval = latestTime - pastReport.created;
    
    // 如果实际时间间隔小于所需的间隔减去容忍度，则认为数据不足
    return actualInterval >= intervalSeconds - timeDeltaTolerance;
}


export default function NMTChart() {
    const [nmt, setNMT] = useState<NMTReport[]>([]);
    const [allCategories, setAllCategories] = useState<string[]>([]);
    const [selectedCategories, setSelectedCategories] = useState<string[]>(["Total"]);
    const [showReserved, setShowReserved] = useState<boolean>(false);
    
    const [sortCriteria, setSortCriteria] = useState<SortCriteria>('none');
    const [sortType, setSortType] = useState<SortType>('delta'); // 新增排序类型状态
    
    // 新增: 时间范围状态和触发器
    const defaultEndTime = Math.floor(Date.now() / 1000); // 默认结束时间为当前时间
    const defaultStartTime = defaultEndTime - 3600; // 默认开始时间为当前时间前 1 小时 (与 1h 快速按钮对应)
    
    const [startTime, setStartTime] = useState<number>(defaultStartTime);
    const [endTime, setEndTime] = useState<number>(defaultEndTime);
    const [fetchTrigger, setFetchTrigger] = useState<number>(0); // 用于触发 useEffect

    // 格式化时间戳为 ISO 字符串，用于输入框显示
    const timestampToDateTimeLocal = (timestamp: number): string => {
        const date = new Date(timestamp * 1000);
        const pad = (n: number) => n.toString().padStart(2, '0');
        
        const year = date.getFullYear();
        const month = pad(date.getMonth() + 1);
        const day = pad(date.getDate());
        const hours = pad(date.getHours());
        const minutes = pad(date.getMinutes());
        
        // 注意：不包含秒，因为 datetime-local 输入框通常不显示秒
        return `${year}-${month}-${day}T${hours}:${minutes}`;
    };

    // 将 datetime-local 字符串转换为时间戳 (秒)
    const dateTimeLocalToTimestamp = (datetimeLocal: string): number => {
        // 假设输入的时间是本地时间
        const date = new Date(datetimeLocal);
        // 转换成秒级时间戳
        return Math.floor(date.getTime() / 1000);
    };

    // 处理快速时间范围选择
    const handleQuickRange = (seconds: number) => {
        const now = Math.floor(Date.now() / 1000);
        setEndTime(now);
        setStartTime(now - seconds);
        // 触发数据加载
        setFetchTrigger(prev => prev + 1);
    };

    // --- 1. 数据获取 ---
    useEffect(() => {
        const apiUrl = `/api/nmt?begin=${startTime}&end=${endTime}`;

        fetch(apiUrl, {
            method: "GET",
        })
            .then(res => res.json())
            .then(data => {
                if (!data.data) {
                    setNMT([]);
                    setAllCategories([]);
                    return;
                }

                data.data.sort((a: { created: number; }, b: { created: number; }) => a.created - b.created);
                setNMT(data.data);

                if (data.data.length > 0) {
                    const cats = data.data[0].nmt_entries.map((entry: { name: string; }) => entry.name);
                    setAllCategories(cats);
                    if (!selectedCategories.includes("Total") && cats.includes("Total")) {
                         setSelectedCategories(["Total"]);
                    } else if (selectedCategories.length === 0 && cats.length > 0) {
                         setSelectedCategories([cats.includes("Total") ? "Total" : cats[0]]);
                    }
                } else {
                    setAllCategories([]);
                    setSelectedCategories([]);
                }
            })
            .catch(error => {
                console.error("Error fetching NMT data:", error);
                setNMT([]);
            });
    }, [fetchTrigger]);


    // --- 2. 颜色固定逻辑 (不变) ---
    const colorMap = useMemo(() => {
        const map: Record<string, string> = {};
        allCategories.forEach((cat, index) => {
            map[cat] = PALETTE[index % PALETTE.length];
        });
        return map;
    }, [allCategories]);

    // --- 3. 排序逻辑 (不变) ---
    const sortedCategoriesWithDelta = useMemo((): CategoryListItem[] => {
        if (sortCriteria === 'none' || nmt.length < 2) {
            return allCategories.map(cat => ({ name: cat, delta: null, rate: null }));
        }

        const intervalSeconds = SORT_INTERVALS[sortCriteria];
        
        if (!isDataSufficient(nmt, intervalSeconds)) {
             return allCategories.map(cat => ({ name: cat, delta: null, rate: null })); 
        }

        const latestReport = nmt[nmt.length - 1];
        const latestTime = latestReport.created;
        const targetTime = latestTime - intervalSeconds;
        
        let pastReport = nmt.reduce((prev, curr) => 
            (Math.abs(curr.created - targetTime) < Math.abs(prev.created - targetTime) ? curr : prev), 
            nmt[0]
        );
        
        const actualInterval = latestTime - pastReport.created;
        
        const deltaList: CategoryListItem[] = [];
        
        allCategories.forEach(cat => {
            const latestEntry = latestReport.nmt_entries.find(e => e.name === cat);
            const pastEntry = pastReport.nmt_entries.find(e => e.name === cat);

            const latestCommitted = latestEntry?.committed || 0;
            const pastCommitted = pastEntry?.committed || 0;

            const deltaMB = (latestCommitted - pastCommitted) / 1024;
            const rateMBps = actualInterval > 0 ? deltaMB / actualInterval : 0;
            
            deltaList.push({ 
                name: cat, 
                delta: deltaMB, 
                rate: rateMBps 
            });
        });

        return deltaList.sort((a, b) => {
            if (sortType === 'delta') {
                return (b.delta || 0) - (a.delta || 0);
            } else { // 'rate'
                return (b.rate || 0) - (a.rate || 0);
            }
        });
        
    }, [nmt, allCategories, sortCriteria, sortType]);

    // --- 4. 数据转换和格式化 (不变) ---
    const chartData = useMemo(() => {
        return nmt.map((item) => {
            const dataPoint: any = { created: item.created };
            item.nmt_entries.forEach((entry) => {
                if (allCategories.includes(entry.name)) {
                    dataPoint[`${entry.name}_committed`] = (entry.committed || 0) / 1024;
                    dataPoint[`${entry.name}_reserved`] = (entry.reserved || 0) / 1024;
                }
            });
            return dataPoint;
        });
    }, [nmt, allCategories]);

    const toggleCategory = (category: string) => {
        setSelectedCategories(prev => {
            if (prev.includes(category)) {
                if (prev.length === 1) return prev; 
                return prev.filter(c => c !== category);
            } else {
                return [...prev, category];
            }
        });
    };

    const formatTime = (timestamp: number) => new Date(timestamp * 1000).toLocaleTimeString();
    
    const isDataSufficientForCurrentSort = sortCriteria === 'none' || isDataSufficient(nmt, SORT_INTERVALS[sortCriteria]);
    const isCurrentSortSuccessful = sortCriteria !== 'none' && isDataSufficientForCurrentSort;
    const totalDataPoints = nmt.length;


    const formatDeltaValue = (item: CategoryListItem): string => {
        if (item.delta === null || item.rate === null) return '';

        let value, unit;
        if (sortType === 'delta') {
            value = item.delta;
            unit = ' MB';
        } else { // 'rate'
            value = item.rate;
            unit = ' MB/s';
        }

        const sign = value >= 0 ? '+' : '';
        const precision = sortType === 'rate' && Math.abs(value) < 1 ? 4 : 2; 

        return `${sign}${value.toFixed(precision)}${unit}`;
    };


    return (
        <div style={{ width: '100%', maxWidth: '1918px', height: '95vh', display: 'flex', flexDirection: 'row', gap: '15px', padding: '10px', fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, "Noto Sans", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji"' }}>
            
            {/* 左侧：图表区 (不变) */}
            <div style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column', gap: '10px' }}>
                <div style={{ flex: 1, minHeight: 0 }}>
                    <ResponsiveContainer width="100%" height="100%">
                        <LineChart data={chartData} margin={{ top: 5, right: 30, left: 10, bottom: 5 }}>
                            <XAxis dataKey="created" tickFormatter={formatTime} />
                            <YAxis label={{ value: 'Memory (MB)', angle: -90, position: 'insideLeft' }} />
                            <Tooltip 
                                labelFormatter={(label) => new Date(label * 1000).toLocaleString()}
                                formatter={(value: number, name: string) => {
                                    const [cat, type] = name.split('_');
                                    if (selectedCategories.includes(cat)) {
                                         return [`${value.toFixed(2)} MB`, `${cat} (${type})`];
                                    }
                                    return [value, name];
                                }}
                            />
                            <Legend />
                            
                            {selectedCategories.map((cat) => {
                                const color = colorMap[cat] || "#000";
                                return (
                                    <g key={cat}>
                                        <Line
                                            type="monotone"
                                            dataKey={`${cat}_committed`}
                                            name={`${cat} (Committed)`}
                                            stroke={color}
                                            strokeWidth={2}
                                            dot={false}
                                            isAnimationActive={false}
                                        />
                                        {showReserved && (
                                            <Line
                                                type="monotone"
                                                dataKey={`${cat}_reserved`}
                                                name={`${cat} (Reserved)`}
                                                stroke={color}
                                                strokeWidth={2}
                                                strokeDasharray="5 5"
                                                dot={false}
                                                isAnimationActive={false}
                                            />
                                        )}
                                    </g>
                                );
                            })}
                        </LineChart>
                    </ResponsiveContainer>
                </div>
            </div>

            {/* 右侧：Category 选择和排序控制 (已修改布局) */}
            <div style={{ width: '380px', display: 'flex', flexDirection: 'column', borderLeft: '1px solid #eee', paddingLeft: '15px', overflowY: 'auto' }}>
                
                {/* 1. Show Reserved | 快速范围选择 | 自定义时间范围 - 组合块 */}
                <div style={{ marginBottom: '15px', paddingBottom: '10px', borderBottom: '1px solid #eee' }}>
                    
                    {/* 1a. Show Reserved */}
                    <div style={{ marginBottom: '10px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <label style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '5px' }}>
                            <input 
                                type="checkbox" 
                                checked={showReserved} 
                                onChange={(e) => setShowReserved(e.target.checked)} 
                            />
                            <span style={{ fontWeight: 'bold' }}>Show Reserved (Dashed)</span>
                        </label>
                    </div>

                    {/* 1b. 快速范围选择按钮 */}
                    <span style={{ fontWeight: 'bold', display: 'block', marginBottom: '5px', marginTop: '5px' }}>Quick Range:</span>
                    <div style={{ display: 'flex', gap: '8px', marginBottom: '10px' }}>
                        {Object.keys(QUICK_RANGES).map(key => (
                            <button
                                key={key}
                                onClick={() => handleQuickRange(QUICK_RANGES[key as keyof typeof QUICK_RANGES])}
                                style={{ 
                                    padding: '5px 8px', 
                                    fontSize: '12px',
                                    backgroundColor: '#eee', 
                                    border: '1px solid #ccc', 
                                    borderRadius: '4px', 
                                    cursor: 'pointer' 
                                }}
                            >
                                Last {key}
                            </button>
                        ))}
                    </div>

                    {/* 1c. Custom Time Range */}
                    <span style={{ fontWeight: 'bold', display: 'block', marginBottom: '5px' }}>Custom Time Range:</span>
                    <div style={{ display: 'flex', gap: '5px', marginBottom: '8px' }}>
                        {/* Start Time */}
                        <div style={{ flex: 1, minWidth: 0, fontSize: '14px' }}>
                            <label>Start:</label>
                            <input
                                type="datetime-local"
                                value={timestampToDateTimeLocal(startTime)}
                                onChange={(e) => setStartTime(dateTimeLocalToTimestamp(e.target.value))}
                                style={{ width: '100%', padding: '3px', marginTop: '2px', boxSizing: 'border-box' }}
                            />
                        </div>
                        {/* End Time */}
                        <div style={{ flex: 1, minWidth: 0, fontSize: '14px' }}>
                            <label>End:</label>
                            <input
                                type="datetime-local"
                                value={timestampToDateTimeLocal(endTime)}
                                onChange={(e) => setEndTime(dateTimeLocalToTimestamp(e.target.value))}
                                style={{ width: '100%', padding: '3px', marginTop: '2px', boxSizing: 'border-box' }}
                            />
                        </div>
                    </div>
                    
                    {/* Load Button */}
                    <button 
                        onClick={() => setFetchTrigger(prev => prev + 1)}
                        style={{ padding: '5px', width: '100%', backgroundColor: '#1f77b4', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', marginTop: '5px' }}
                    >
                        Load Data ({totalDataPoints} points currently)
                    </button>
                </div>

                {/* 2. 分类排序控制块 */}
                <div style={{ marginBottom: '15px', paddingBottom: '10px', borderBottom: '1px solid #eee' }}>
                    <span style={{ fontWeight: 'bold', display: 'block', marginBottom: '5px' }}>Sort Criteria:</span>
                    
                    <div style={{ display: 'flex', gap: '15px', marginBottom: '10px' }}>
                        <label style={{ fontSize: '14px', cursor: 'pointer' }}>
                            <input 
                                type="radio" 
                                value="delta" 
                                checked={sortType === 'delta'} 
                                onChange={() => setSortType('delta')}
                            />
                            Absolute Delta
                        </label>
                        <label style={{ fontSize: '14px', cursor: 'pointer' }}>
                            <input 
                                type="radio" 
                                value="rate" 
                                checked={sortType === 'rate'} 
                                onChange={() => setSortType('rate')}
                            />
                            Rate of Change
                        </label>
                    </div>

                    <select 
                        value={sortCriteria} 
                        onChange={(e) => setSortCriteria(e.target.value as SortCriteria)}
                        style={{ width: '100%', padding: '5px' }}
                    >
                        <option value="none">None (Alphabetical)</option>
                        {Object.keys(SORT_INTERVALS).map(key => {
                            const intervalSeconds = SORT_INTERVALS[key as keyof typeof SORT_INTERVALS];
                            const isDisabled = totalDataPoints < 2 || !isDataSufficient(nmt, intervalSeconds);
                            
                            return (
                                <option 
                                    key={key} 
                                    value={key} 
                                    disabled={isDisabled}
                                    title={isDisabled ? `Data interval is less than ${key}.` : `Sort by ${sortType} over the last ${key}.`}
                                >
                                    {key} {sortType === 'delta' ? 'Delta' : 'Rate'} (Largest First)
                                </option>
                            );
                        })}
                    </select>
                    {sortCriteria !== 'none' && (
                        <small style={{ color: isDataSufficientForCurrentSort ? '#2ca02c' : 'red', marginTop: '5px', display: 'block' }}>
                            {isDataSufficientForCurrentSort
                                ? `Categories sorted by ${sortType === 'delta' ? 'absolute memory change' : 'change rate (MB/s)'}.` 
                                : `Data insufficient for ${sortCriteria} interval. Displaying alphabetical order.`}
                        </small>
                    )}
                </div>
                
                {/* 3. 分类 Checkbox 列表 */}
                <div style={{ fontWeight: 'bold', marginBottom: '8px' }}>Categories:</div>
                <div style={{ display: 'flex', flexDirection: 'column', gap: '5px' }}>
                    {sortedCategoriesWithDelta.map((item) => (
                        <label 
                            key={item.name} 
                            style={{ 
                                display: 'flex', 
                                alignItems: 'center', 
                                gap: '4px', 
                                cursor: 'pointer', 
                                fontSize: '14px', 
                                justifyContent: 'space-between'
                            }}
                        >
                            <div style={{ display: 'flex', alignItems: 'center', minWidth: 0 }}>
                                <input
                                    type="checkbox"
                                    checked={selectedCategories.includes(item.name)}
                                    onChange={() => toggleCategory(item.name)}
                                />
                                <span style={{ 
                                    color: colorMap[item.name] || '#666', 
                                    whiteSpace: 'nowrap',
                                    overflow: 'hidden',
                                    textOverflow: 'ellipsis'
                                }}>
                                    {item.name}
                                </span>
                            </div>
                            
                            {isCurrentSortSuccessful && (item.delta !== null || item.rate !== null) && (
                                <span style={{ 
                                    marginLeft: '10px', 
                                    fontSize: '12px', 
                                    fontWeight: 'bold', 
                                    color: (sortType === 'delta' ? item.delta! : item.rate!) >= 0 ? '#d62728' : '#2ca02c', 
                                    flexShrink: 0
                                }}>
                                    {formatDeltaValue(item)}
                                </span>
                            )}
                        </label>
                    ))}
                </div>
            </div>
        </div>
    );
}