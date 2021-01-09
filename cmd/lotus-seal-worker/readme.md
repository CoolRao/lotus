#### 配置文件参数说明


    {
      "WorkerName": "",                 // worker主机名称
      "AddPieceMax": 3,                 // 同时做AP最大数量 
      "PreCommit1Max": 7,               // 同时做P1最大数量
      "PreCommit2Max": 1,               // 同时做P2最大数量
      "Commit2Max": 1,                  // 同时做C2最大数量    
      "DiskHoldMax": 0,                 // 
      "APDiskHoldMax": 0,
      "ForceP1FromLocalAP": false,      // P1只做本地AP过来任务   
      "ForceP2FromLocalP1": false,      // P2只做本地P1过来任务
      "ForceC2FromLocalP2": false,      // C2只做本地P2过来任务
      "IsPlanOffline": false,           // 自动下线
      "AllowP2C2Parallel": false,       // P2/C2并行
      "IgnoreOutOfSpace": false         // 忽略存储空间提示
    }