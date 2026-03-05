// 英语常用固定搭配词库
// 用于识别单词在上下文中是否形成固定搭配

export interface Collocation {
  pattern: string;       // 搭配模式，使用 {word} 表示核心词位置
  phrase: string;        // 完整短语示例
  meaning: string;       // 中文含义
  keywords: string[];    // 触发关键词（当用户选中这些词时检测）
}

// 常用动词搭配
const verbCollocations: Collocation[] = [
  // make
  { pattern: 'make a decision', phrase: 'make a decision', meaning: '做决定', keywords: ['make', 'decision'] },
  { pattern: 'make a difference', phrase: 'make a difference', meaning: '产生影响', keywords: ['make', 'difference'] },
  { pattern: 'make a mistake', phrase: 'make a mistake', meaning: '犯错误', keywords: ['make', 'mistake'] },
  { pattern: 'make an effort', phrase: 'make an effort', meaning: '努力', keywords: ['make', 'effort'] },
  { pattern: 'make progress', phrase: 'make progress', meaning: '取得进步', keywords: ['make', 'progress'] },
  { pattern: 'make sense', phrase: 'make sense', meaning: '有意义，讲得通', keywords: ['make', 'sense'] },
  { pattern: 'make sure', phrase: 'make sure', meaning: '确保', keywords: ['make', 'sure'] },
  { pattern: 'make up', phrase: 'make up', meaning: '编造；化妆；弥补', keywords: ['make', 'up'] },
  { pattern: 'make up for', phrase: 'make up for', meaning: '弥补', keywords: ['make', 'up', 'for'] },
  { pattern: 'make friends', phrase: 'make friends', meaning: '交朋友', keywords: ['make', 'friends'] },
  { pattern: 'make money', phrase: 'make money', meaning: '赚钱', keywords: ['make', 'money'] },
  
  // take
  { pattern: 'take a look', phrase: 'take a look', meaning: '看一看', keywords: ['take', 'look'] },
  { pattern: 'take a break', phrase: 'take a break', meaning: '休息一下', keywords: ['take', 'break'] },
  { pattern: 'take a chance', phrase: 'take a chance', meaning: '冒险', keywords: ['take', 'chance'] },
  { pattern: 'take action', phrase: 'take action', meaning: '采取行动', keywords: ['take', 'action'] },
  { pattern: 'take advantage of', phrase: 'take advantage of', meaning: '利用', keywords: ['take', 'advantage'] },
  { pattern: 'take care of', phrase: 'take care of', meaning: '照顾', keywords: ['take', 'care'] },
  { pattern: 'take into account', phrase: 'take into account', meaning: '考虑到', keywords: ['take', 'account'] },
  { pattern: 'take it easy', phrase: 'take it easy', meaning: '放轻松', keywords: ['take', 'easy'] },
  { pattern: 'take notes', phrase: 'take notes', meaning: '做笔记', keywords: ['take', 'notes'] },
  { pattern: 'take part in', phrase: 'take part in', meaning: '参加', keywords: ['take', 'part'] },
  { pattern: 'take place', phrase: 'take place', meaning: '发生', keywords: ['take', 'place'] },
  { pattern: 'take responsibility', phrase: 'take responsibility', meaning: '承担责任', keywords: ['take', 'responsibility'] },
  { pattern: 'take time', phrase: 'take time', meaning: '花时间', keywords: ['take', 'time'] },
  { pattern: 'take turns', phrase: 'take turns', meaning: '轮流', keywords: ['take', 'turns'] },
  { pattern: 'take off', phrase: 'take off', meaning: '起飞；脱下', keywords: ['take', 'off'] },
  { pattern: 'take over', phrase: 'take over', meaning: '接管', keywords: ['take', 'over'] },
  
  // get
  { pattern: 'get rid of', phrase: 'get rid of', meaning: '摆脱', keywords: ['get', 'rid'] },
  { pattern: 'get along with', phrase: 'get along with', meaning: '与...相处', keywords: ['get', 'along'] },
  { pattern: 'get in touch', phrase: 'get in touch', meaning: '取得联系', keywords: ['get', 'touch'] },
  { pattern: 'get used to', phrase: 'get used to', meaning: '习惯于', keywords: ['get', 'used'] },
  { pattern: 'get ready', phrase: 'get ready', meaning: '准备好', keywords: ['get', 'ready'] },
  { pattern: 'get started', phrase: 'get started', meaning: '开始', keywords: ['get', 'started'] },
  { pattern: 'get better', phrase: 'get better', meaning: '变好', keywords: ['get', 'better'] },
  { pattern: 'get worse', phrase: 'get worse', meaning: '变糟', keywords: ['get', 'worse'] },
  { pattern: 'get through', phrase: 'get through', meaning: '度过；完成', keywords: ['get', 'through'] },
  { pattern: 'get over', phrase: 'get over', meaning: '克服', keywords: ['get', 'over'] },
  
  // give
  { pattern: 'give up', phrase: 'give up', meaning: '放弃', keywords: ['give', 'up'] },
  { pattern: 'give in', phrase: 'give in', meaning: '屈服', keywords: ['give', 'in'] },
  { pattern: 'give away', phrase: 'give away', meaning: '赠送；泄露', keywords: ['give', 'away'] },
  { pattern: 'give back', phrase: 'give back', meaning: '归还', keywords: ['give', 'back'] },
  { pattern: 'give rise to', phrase: 'give rise to', meaning: '导致', keywords: ['give', 'rise'] },
  
  // have
  { pattern: 'have access to', phrase: 'have access to', meaning: '有权使用', keywords: ['have', 'access'] },
  { pattern: 'have a good time', phrase: 'have a good time', meaning: '玩得开心', keywords: ['have', 'good', 'time'] },
  { pattern: 'have an impact on', phrase: 'have an impact on', meaning: '对...有影响', keywords: ['have', 'impact'] },
  { pattern: 'have difficulty', phrase: 'have difficulty', meaning: '有困难', keywords: ['have', 'difficulty'] },
  { pattern: 'have in common', phrase: 'have in common', meaning: '有共同点', keywords: ['have', 'common'] },
  { pattern: 'have no idea', phrase: 'have no idea', meaning: '不知道', keywords: ['have', 'idea'] },
  { pattern: 'have nothing to do with', phrase: 'have nothing to do with', meaning: '与...无关', keywords: ['have', 'nothing', 'do'] },
  
  // do
  { pattern: 'do good', phrase: 'do good', meaning: '做好事', keywords: ['do', 'good'] },
  { pattern: 'do harm', phrase: 'do harm', meaning: '造成伤害', keywords: ['do', 'harm'] },
  { pattern: 'do one\'s best', phrase: 'do one\'s best', meaning: '尽力', keywords: ['do', 'best'] },
  { pattern: 'do the dishes', phrase: 'do the dishes', meaning: '洗碗', keywords: ['do', 'dishes'] },
  { pattern: 'do homework', phrase: 'do homework', meaning: '做作业', keywords: ['do', 'homework'] },
  { pattern: 'do business', phrase: 'do business', meaning: '做生意', keywords: ['do', 'business'] },
  { pattern: 'do away with', phrase: 'do away with', meaning: '废除', keywords: ['do', 'away'] },
  
  // come
  { pattern: 'come across', phrase: 'come across', meaning: '偶遇', keywords: ['come', 'across'] },
  { pattern: 'come up with', phrase: 'come up with', meaning: '想出', keywords: ['come', 'up', 'with'] },
  { pattern: 'come to', phrase: 'come to', meaning: '达到；苏醒', keywords: ['come', 'to'] },
  { pattern: 'come true', phrase: 'come true', meaning: '实现', keywords: ['come', 'true'] },
  { pattern: 'come back', phrase: 'come back', meaning: '回来', keywords: ['come', 'back'] },
  { pattern: 'come from', phrase: 'come from', meaning: '来自', keywords: ['come', 'from'] },
  { pattern: 'come along', phrase: 'come along', meaning: '一起来；进展', keywords: ['come', 'along'] },
  
  // go
  { pattern: 'go ahead', phrase: 'go ahead', meaning: '继续；开始吧', keywords: ['go', 'ahead'] },
  { pattern: 'go on', phrase: 'go on', meaning: '继续', keywords: ['go', 'on'] },
  { pattern: 'go through', phrase: 'go through', meaning: '经历', keywords: ['go', 'through'] },
  { pattern: 'go over', phrase: 'go over', meaning: '复习；检查', keywords: ['go', 'over'] },
  { pattern: 'go wrong', phrase: 'go wrong', meaning: '出错', keywords: ['go', 'wrong'] },
  { pattern: 'go out', phrase: 'go out', meaning: '外出；熄灭', keywords: ['go', 'out'] },
  { pattern: 'go back', phrase: 'go back', meaning: '回去', keywords: ['go', 'back'] },
  
  // put
  { pattern: 'put off', phrase: 'put off', meaning: '推迟', keywords: ['put', 'off'] },
  { pattern: 'put on', phrase: 'put on', meaning: '穿上', keywords: ['put', 'on'] },
  { pattern: 'put up with', phrase: 'put up with', meaning: '忍受', keywords: ['put', 'up', 'with'] },
  { pattern: 'put forward', phrase: 'put forward', meaning: '提出', keywords: ['put', 'forward'] },
  { pattern: 'put into practice', phrase: 'put into practice', meaning: '付诸实践', keywords: ['put', 'practice'] },
  { pattern: 'put emphasis on', phrase: 'put emphasis on', meaning: '强调', keywords: ['put', 'emphasis'] },
  
  // turn
  { pattern: 'turn out', phrase: 'turn out', meaning: '结果是', keywords: ['turn', 'out'] },
  { pattern: 'turn off', phrase: 'turn off', meaning: '关闭', keywords: ['turn', 'off'] },
  { pattern: 'turn on', phrase: 'turn on', meaning: '打开', keywords: ['turn', 'on'] },
  { pattern: 'turn around', phrase: 'turn around', meaning: '转身；扭转', keywords: ['turn', 'around'] },
  { pattern: 'turn down', phrase: 'turn down', meaning: '拒绝；调低', keywords: ['turn', 'down'] },
  { pattern: 'turn up', phrase: 'turn up', meaning: '出现；调高', keywords: ['turn', 'up'] },
  { pattern: 'turn into', phrase: 'turn into', meaning: '变成', keywords: ['turn', 'into'] },
  
  // look
  { pattern: 'look after', phrase: 'look after', meaning: '照顾', keywords: ['look', 'after'] },
  { pattern: 'look at', phrase: 'look at', meaning: '看', keywords: ['look', 'at'] },
  { pattern: 'look for', phrase: 'look for', meaning: '寻找', keywords: ['look', 'for'] },
  { pattern: 'look forward to', phrase: 'look forward to', meaning: '期待', keywords: ['look', 'forward'] },
  { pattern: 'look into', phrase: 'look into', meaning: '调查', keywords: ['look', 'into'] },
  { pattern: 'look like', phrase: 'look like', meaning: '看起来像', keywords: ['look', 'like'] },
  { pattern: 'look up', phrase: 'look up', meaning: '查阅', keywords: ['look', 'up'] },
  { pattern: 'look up to', phrase: 'look up to', meaning: '尊敬', keywords: ['look', 'up', 'to'] },
  { pattern: 'look down on', phrase: 'look down on', meaning: '看不起', keywords: ['look', 'down'] },
  
  // set
  { pattern: 'set up', phrase: 'set up', meaning: '建立', keywords: ['set', 'up'] },
  { pattern: 'set off', phrase: 'set off', meaning: '出发；引发', keywords: ['set', 'off'] },
  { pattern: 'set out', phrase: 'set out', meaning: '出发；开始', keywords: ['set', 'out'] },
  { pattern: 'set aside', phrase: 'set aside', meaning: '留出', keywords: ['set', 'aside'] },
  
  // bring
  { pattern: 'bring about', phrase: 'bring about', meaning: '引起', keywords: ['bring', 'about'] },
  { pattern: 'bring up', phrase: 'bring up', meaning: '抚养；提出', keywords: ['bring', 'up'] },
  { pattern: 'bring back', phrase: 'bring back', meaning: '带回', keywords: ['bring', 'back'] },
  { pattern: 'bring out', phrase: 'bring out', meaning: '使显现', keywords: ['bring', 'out'] },
  
  // break
  { pattern: 'break down', phrase: 'break down', meaning: '分解；崩溃', keywords: ['break', 'down'] },
  { pattern: 'break up', phrase: 'break up', meaning: '分手；解散', keywords: ['break', 'up'] },
  { pattern: 'break out', phrase: 'break out', meaning: '爆发', keywords: ['break', 'out'] },
  { pattern: 'break through', phrase: 'break through', meaning: '突破', keywords: ['break', 'through'] },
  
  // carry
  { pattern: 'carry out', phrase: 'carry out', meaning: '执行', keywords: ['carry', 'out'] },
  { pattern: 'carry on', phrase: 'carry on', meaning: '继续', keywords: ['carry', 'on'] },
  
  // work
  { pattern: 'work out', phrase: 'work out', meaning: '解决；锻炼', keywords: ['work', 'out'] },
  { pattern: 'work on', phrase: 'work on', meaning: '从事', keywords: ['work', 'on'] },
  
  // keep
  { pattern: 'keep up with', phrase: 'keep up with', meaning: '跟上', keywords: ['keep', 'up'] },
  { pattern: 'keep in mind', phrase: 'keep in mind', meaning: '记住', keywords: ['keep', 'mind'] },
  { pattern: 'keep an eye on', phrase: 'keep an eye on', meaning: '留意', keywords: ['keep', 'eye'] },
  { pattern: 'keep in touch', phrase: 'keep in touch', meaning: '保持联系', keywords: ['keep', 'touch'] },
  
  // pay
  { pattern: 'pay attention to', phrase: 'pay attention to', meaning: '注意', keywords: ['pay', 'attention'] },
  { pattern: 'pay off', phrase: 'pay off', meaning: '还清；取得成功', keywords: ['pay', 'off'] },
  
  // run
  { pattern: 'run out of', phrase: 'run out of', meaning: '用完', keywords: ['run', 'out'] },
  { pattern: 'run into', phrase: 'run into', meaning: '遇到', keywords: ['run', 'into'] },
  
  // pick
  { pattern: 'pick up', phrase: 'pick up', meaning: '捡起；学会', keywords: ['pick', 'up'] },
  { pattern: 'pick out', phrase: 'pick out', meaning: '挑选', keywords: ['pick', 'out'] },
  
  // point
  { pattern: 'point out', phrase: 'point out', meaning: '指出', keywords: ['point', 'out'] },
  
  // figure
  { pattern: 'figure out', phrase: 'figure out', meaning: '弄清楚', keywords: ['figure', 'out'] },
  
  // find
  { pattern: 'find out', phrase: 'find out', meaning: '发现', keywords: ['find', 'out'] },
  
  // fill
  { pattern: 'fill in', phrase: 'fill in', meaning: '填写', keywords: ['fill', 'in'] },
  { pattern: 'fill out', phrase: 'fill out', meaning: '填写（表格）', keywords: ['fill', 'out'] },
  
  // hold
  { pattern: 'hold on', phrase: 'hold on', meaning: '等等；坚持', keywords: ['hold', 'on'] },
  { pattern: 'hold back', phrase: 'hold back', meaning: '阻止', keywords: ['hold', 'back'] },
  
  // hand
  { pattern: 'hand in', phrase: 'hand in', meaning: '上交', keywords: ['hand', 'in'] },
  { pattern: 'hand out', phrase: 'hand out', meaning: '分发', keywords: ['hand', 'out'] },
  
  // check
  { pattern: 'check in', phrase: 'check in', meaning: '登记入住', keywords: ['check', 'in'] },
  { pattern: 'check out', phrase: 'check out', meaning: '结账离开；查看', keywords: ['check', 'out'] },
  
  // call
  { pattern: 'call off', phrase: 'call off', meaning: '取消', keywords: ['call', 'off'] },
  { pattern: 'call on', phrase: 'call on', meaning: '拜访；号召', keywords: ['call', 'on'] },
  { pattern: 'call for', phrase: 'call for', meaning: '要求', keywords: ['call', 'for'] },
];

// 常用介词搭配
const prepositionCollocations: Collocation[] = [
  // in
  { pattern: 'in fact', phrase: 'in fact', meaning: '事实上', keywords: ['in', 'fact'] },
  { pattern: 'in general', phrase: 'in general', meaning: '总的来说', keywords: ['in', 'general'] },
  { pattern: 'in particular', phrase: 'in particular', meaning: '尤其', keywords: ['in', 'particular'] },
  { pattern: 'in addition', phrase: 'in addition', meaning: '此外', keywords: ['in', 'addition'] },
  { pattern: 'in conclusion', phrase: 'in conclusion', meaning: '总之', keywords: ['in', 'conclusion'] },
  { pattern: 'in contrast', phrase: 'in contrast', meaning: '相比之下', keywords: ['in', 'contrast'] },
  { pattern: 'in spite of', phrase: 'in spite of', meaning: '尽管', keywords: ['in', 'spite'] },
  { pattern: 'in terms of', phrase: 'in terms of', meaning: '就...而言', keywords: ['in', 'terms'] },
  { pattern: 'in case of', phrase: 'in case of', meaning: '万一', keywords: ['in', 'case'] },
  { pattern: 'in favor of', phrase: 'in favor of', meaning: '支持', keywords: ['in', 'favor'] },
  { pattern: 'in charge of', phrase: 'in charge of', meaning: '负责', keywords: ['in', 'charge'] },
  { pattern: 'in the end', phrase: 'in the end', meaning: '最后', keywords: ['in', 'end'] },
  { pattern: 'in other words', phrase: 'in other words', meaning: '换句话说', keywords: ['in', 'other', 'words'] },
  { pattern: 'in order to', phrase: 'in order to', meaning: '为了', keywords: ['in', 'order'] },
  { pattern: 'in the meantime', phrase: 'in the meantime', meaning: '与此同时', keywords: ['in', 'meantime'] },
  { pattern: 'in the long run', phrase: 'in the long run', meaning: '从长远来看', keywords: ['in', 'long', 'run'] },
  
  // on
  { pattern: 'on purpose', phrase: 'on purpose', meaning: '故意', keywords: ['on', 'purpose'] },
  { pattern: 'on time', phrase: 'on time', meaning: '准时', keywords: ['on', 'time'] },
  { pattern: 'on the other hand', phrase: 'on the other hand', meaning: '另一方面', keywords: ['on', 'other', 'hand'] },
  { pattern: 'on the contrary', phrase: 'on the contrary', meaning: '相反', keywords: ['on', 'contrary'] },
  { pattern: 'on behalf of', phrase: 'on behalf of', meaning: '代表', keywords: ['on', 'behalf'] },
  { pattern: 'on account of', phrase: 'on account of', meaning: '由于', keywords: ['on', 'account'] },
  { pattern: 'on average', phrase: 'on average', meaning: '平均', keywords: ['on', 'average'] },
  
  // at
  { pattern: 'at first', phrase: 'at first', meaning: '起初', keywords: ['at', 'first'] },
  { pattern: 'at last', phrase: 'at last', meaning: '最后', keywords: ['at', 'last'] },
  { pattern: 'at least', phrase: 'at least', meaning: '至少', keywords: ['at', 'least'] },
  { pattern: 'at most', phrase: 'at most', meaning: '最多', keywords: ['at', 'most'] },
  { pattern: 'at once', phrase: 'at once', meaning: '立刻', keywords: ['at', 'once'] },
  { pattern: 'at present', phrase: 'at present', meaning: '目前', keywords: ['at', 'present'] },
  { pattern: 'at the same time', phrase: 'at the same time', meaning: '同时', keywords: ['at', 'same', 'time'] },
  { pattern: 'at all costs', phrase: 'at all costs', meaning: '不惜一切代价', keywords: ['at', 'all', 'costs'] },
  
  // by
  { pattern: 'by the way', phrase: 'by the way', meaning: '顺便说一下', keywords: ['by', 'way'] },
  { pattern: 'by all means', phrase: 'by all means', meaning: '当然可以', keywords: ['by', 'all', 'means'] },
  { pattern: 'by no means', phrase: 'by no means', meaning: '绝不', keywords: ['by', 'no', 'means'] },
  { pattern: 'by accident', phrase: 'by accident', meaning: '偶然', keywords: ['by', 'accident'] },
  { pattern: 'by chance', phrase: 'by chance', meaning: '偶然', keywords: ['by', 'chance'] },
  { pattern: 'by far', phrase: 'by far', meaning: '到目前为止最', keywords: ['by', 'far'] },
  
  // for
  { pattern: 'for example', phrase: 'for example', meaning: '例如', keywords: ['for', 'example'] },
  { pattern: 'for instance', phrase: 'for instance', meaning: '例如', keywords: ['for', 'instance'] },
  { pattern: 'for the sake of', phrase: 'for the sake of', meaning: '为了...的缘故', keywords: ['for', 'sake'] },
  { pattern: 'for good', phrase: 'for good', meaning: '永远', keywords: ['for', 'good'] },
  
  // out
  { pattern: 'out of date', phrase: 'out of date', meaning: '过时的', keywords: ['out', 'date'] },
  { pattern: 'out of order', phrase: 'out of order', meaning: '故障的', keywords: ['out', 'order'] },
  { pattern: 'out of control', phrase: 'out of control', meaning: '失控的', keywords: ['out', 'control'] },
  { pattern: 'out of question', phrase: 'out of question', meaning: '毫无疑问', keywords: ['out', 'question'] },
  { pattern: 'out of the question', phrase: 'out of the question', meaning: '不可能', keywords: ['out', 'the', 'question'] },
  
  // as
  { pattern: 'as a result', phrase: 'as a result', meaning: '结果', keywords: ['as', 'result'] },
  { pattern: 'as a matter of fact', phrase: 'as a matter of fact', meaning: '事实上', keywords: ['as', 'matter', 'fact'] },
  { pattern: 'as well as', phrase: 'as well as', meaning: '和...一样', keywords: ['as', 'well'] },
  { pattern: 'as long as', phrase: 'as long as', meaning: '只要', keywords: ['as', 'long'] },
  { pattern: 'as far as', phrase: 'as far as', meaning: '就...而言', keywords: ['as', 'far'] },
  { pattern: 'as soon as', phrase: 'as soon as', meaning: '一...就', keywords: ['as', 'soon'] },
  { pattern: 'as if', phrase: 'as if', meaning: '好像', keywords: ['as', 'if'] },
  { pattern: 'as though', phrase: 'as though', meaning: '好像', keywords: ['as', 'though'] },
  { pattern: 'as usual', phrase: 'as usual', meaning: '照常', keywords: ['as', 'usual'] },
  
  // with
  { pattern: 'with regard to', phrase: 'with regard to', meaning: '关于', keywords: ['with', 'regard'] },
  { pattern: 'with respect to', phrase: 'with respect to', meaning: '关于', keywords: ['with', 'respect'] },
  
  // to
  { pattern: 'to some extent', phrase: 'to some extent', meaning: '在某种程度上', keywords: ['to', 'extent'] },
  { pattern: 'to a certain degree', phrase: 'to a certain degree', meaning: '在一定程度上', keywords: ['to', 'degree'] },
  { pattern: 'to the point', phrase: 'to the point', meaning: '切题的', keywords: ['to', 'point'] },
  
  // from
  { pattern: 'from time to time', phrase: 'from time to time', meaning: '不时', keywords: ['from', 'time'] },
  { pattern: 'from now on', phrase: 'from now on', meaning: '从现在起', keywords: ['from', 'now'] },
];

// 常用形容词搭配
const adjectiveCollocations: Collocation[] = [
  { pattern: 'be aware of', phrase: 'be aware of', meaning: '意识到', keywords: ['aware', 'of'] },
  { pattern: 'be capable of', phrase: 'be capable of', meaning: '能够', keywords: ['capable', 'of'] },
  { pattern: 'be familiar with', phrase: 'be familiar with', meaning: '熟悉', keywords: ['familiar', 'with'] },
  { pattern: 'be good at', phrase: 'be good at', meaning: '擅长', keywords: ['good', 'at'] },
  { pattern: 'be interested in', phrase: 'be interested in', meaning: '对...感兴趣', keywords: ['interested', 'in'] },
  { pattern: 'be known for', phrase: 'be known for', meaning: '以...著称', keywords: ['known', 'for'] },
  { pattern: 'be responsible for', phrase: 'be responsible for', meaning: '负责', keywords: ['responsible', 'for'] },
  { pattern: 'be similar to', phrase: 'be similar to', meaning: '与...相似', keywords: ['similar', 'to'] },
  { pattern: 'be different from', phrase: 'be different from', meaning: '与...不同', keywords: ['different', 'from'] },
  { pattern: 'be based on', phrase: 'be based on', meaning: '基于', keywords: ['based', 'on'] },
  { pattern: 'be related to', phrase: 'be related to', meaning: '与...有关', keywords: ['related', 'to'] },
  { pattern: 'be associated with', phrase: 'be associated with', meaning: '与...相关', keywords: ['associated', 'with'] },
  { pattern: 'be satisfied with', phrase: 'be satisfied with', meaning: '对...满意', keywords: ['satisfied', 'with'] },
  { pattern: 'be worried about', phrase: 'be worried about', meaning: '担心', keywords: ['worried', 'about'] },
  { pattern: 'be surprised at', phrase: 'be surprised at', meaning: '对...惊讶', keywords: ['surprised', 'at'] },
  { pattern: 'be proud of', phrase: 'be proud of', meaning: '为...自豪', keywords: ['proud', 'of'] },
  { pattern: 'be afraid of', phrase: 'be afraid of', meaning: '害怕', keywords: ['afraid', 'of'] },
  { pattern: 'be fond of', phrase: 'be fond of', meaning: '喜欢', keywords: ['fond', 'of'] },
  { pattern: 'be tired of', phrase: 'be tired of', meaning: '厌倦', keywords: ['tired', 'of'] },
  { pattern: 'be full of', phrase: 'be full of', meaning: '充满', keywords: ['full', 'of'] },
  { pattern: 'be short of', phrase: 'be short of', meaning: '缺少', keywords: ['short', 'of'] },
  { pattern: 'be sure of', phrase: 'be sure of', meaning: '确信', keywords: ['sure', 'of'] },
  { pattern: 'be likely to', phrase: 'be likely to', meaning: '可能', keywords: ['likely', 'to'] },
  { pattern: 'be supposed to', phrase: 'be supposed to', meaning: '应该', keywords: ['supposed', 'to'] },
  { pattern: 'be willing to', phrase: 'be willing to', meaning: '愿意', keywords: ['willing', 'to'] },
  { pattern: 'be able to', phrase: 'be able to', meaning: '能够', keywords: ['able', 'to'] },
  { pattern: 'be about to', phrase: 'be about to', meaning: '即将', keywords: ['about', 'to'] },
];

// 其他常用搭配
const otherCollocations: Collocation[] = [
  // 名词搭配
  { pattern: 'a wide range of', phrase: 'a wide range of', meaning: '广泛的', keywords: ['wide', 'range'] },
  { pattern: 'a variety of', phrase: 'a variety of', meaning: '各种各样的', keywords: ['variety', 'of'] },
  { pattern: 'a number of', phrase: 'a number of', meaning: '许多', keywords: ['number', 'of'] },
  { pattern: 'a great deal of', phrase: 'a great deal of', meaning: '大量的', keywords: ['great', 'deal'] },
  { pattern: 'a series of', phrase: 'a series of', meaning: '一系列', keywords: ['series', 'of'] },
  { pattern: 'a lack of', phrase: 'a lack of', meaning: '缺乏', keywords: ['lack', 'of'] },
  { pattern: 'the majority of', phrase: 'the majority of', meaning: '大多数', keywords: ['majority', 'of'] },
  
  // 连接词搭配
  { pattern: 'not only...but also', phrase: 'not only...but also', meaning: '不仅...而且', keywords: ['not', 'only', 'but', 'also'] },
  { pattern: 'either...or', phrase: 'either...or', meaning: '要么...要么', keywords: ['either', 'or'] },
  { pattern: 'neither...nor', phrase: 'neither...nor', meaning: '既不...也不', keywords: ['neither', 'nor'] },
  { pattern: 'whether...or', phrase: 'whether...or', meaning: '无论...还是', keywords: ['whether', 'or'] },
  { pattern: 'both...and', phrase: 'both...and', meaning: '既...又', keywords: ['both', 'and'] },
  
  // 其他
  { pattern: 'more and more', phrase: 'more and more', meaning: '越来越', keywords: ['more', 'and', 'more'] },
  { pattern: 'sooner or later', phrase: 'sooner or later', meaning: '迟早', keywords: ['sooner', 'later'] },
  { pattern: 'step by step', phrase: 'step by step', meaning: '一步一步', keywords: ['step', 'by', 'step'] },
  { pattern: 'face to face', phrase: 'face to face', meaning: '面对面', keywords: ['face', 'to', 'face'] },
  { pattern: 'side by side', phrase: 'side by side', meaning: '并肩', keywords: ['side', 'by', 'side'] },
  { pattern: 'day by day', phrase: 'day by day', meaning: '日复一日', keywords: ['day', 'by', 'day'] },
  { pattern: 'one by one', phrase: 'one by one', meaning: '一个一个', keywords: ['one', 'by', 'one'] },
  { pattern: 'first of all', phrase: 'first of all', meaning: '首先', keywords: ['first', 'of', 'all'] },
  { pattern: 'after all', phrase: 'after all', meaning: '毕竟', keywords: ['after', 'all'] },
  { pattern: 'above all', phrase: 'above all', meaning: '最重要的是', keywords: ['above', 'all'] },
  { pattern: 'all in all', phrase: 'all in all', meaning: '总的来说', keywords: ['all', 'in', 'all'] },
  { pattern: 'once in a while', phrase: 'once in a while', meaning: '偶尔', keywords: ['once', 'while'] },
  { pattern: 'all of a sudden', phrase: 'all of a sudden', meaning: '突然', keywords: ['all', 'sudden'] },
  { pattern: 'little by little', phrase: 'little by little', meaning: '逐渐', keywords: ['little', 'by', 'little'] },
  { pattern: 'now and then', phrase: 'now and then', meaning: '偶尔', keywords: ['now', 'and', 'then'] },
  { pattern: 'time and again', phrase: 'time and again', meaning: '屡次', keywords: ['time', 'again'] },
];

// 合并所有搭配
export const allCollocations: Collocation[] = [
  ...verbCollocations,
  ...prepositionCollocations,
  ...adjectiveCollocations,
  ...otherCollocations,
];

// 根据选中单词和上下文查找可能的固定搭配
export function findCollocations(word: string, context: string): Collocation[] {
  const lowerWord = word.toLowerCase();
  const lowerContext = context.toLowerCase();
  const foundCollocations: Collocation[] = [];
  
  for (const collocation of allCollocations) {
    // 检查选中的单词是否是该搭配的关键词之一
    if (collocation.keywords.includes(lowerWord)) {
      // 检查上下文中是否包含这个搭配的所有关键词
      const allKeywordsInContext = collocation.keywords.every(keyword => 
        lowerContext.includes(keyword)
      );
      
      // 更严格的检查：搭配短语是否出现在上下文中
      const phraseInContext = lowerContext.includes(collocation.phrase.toLowerCase());
      
      if (phraseInContext || allKeywordsInContext) {
        // 避免重复
        if (!foundCollocations.some(c => c.phrase === collocation.phrase)) {
          foundCollocations.push(collocation);
        }
      }
    }
  }
  
  return foundCollocations;
}
