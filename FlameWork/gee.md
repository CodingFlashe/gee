###### http.Handler

任何实现了接口的类型都能作为处理http请求的实例

![image-20221024230230896](D:\typora\笔记\photo\image-20221024230230896.png) 



###### day1

搭建gee框架原型

1. 定义一个结构体Engine，里面包含一个字段(姑且称为路由映射表)，为map类型，目的是存储不同的请求方法以及静态路由对应的处理方法
2. 定义构造器
3. 定义一个方法：根据请求方法和静态路由生成唯一的路径(方法:把**请求方法和静态路由拼接到一起**，这样子就是唯一的请求路径，如`GET-/hello`)，同时把路径和处理方法添加到路匹映射表
4. 定义 GET、POST方法：调用步骤3定义的方法，传入三个参数，请求方法、静态路由、处理器方法
5. 封装标准库的ListenAndServe方法
6. 实现ServeHTTP方法，通过获取请求体里的请求方法(r.Mathod)和请求的路径(r.URL.Path)，在路由映射表里面查找相应的处理器



###### day2

设计上下文环境(Context)

1. 定义结构体Context封装请求体、请求对应的响应、请求路径、请求路径以及状态码，后面三个是经常要使用的，所以直接封装在结构体中。对应的构造器方法只需要传入请求体和响应体，**请求方法和路径从请求体获取，状态码在后续方法中设置**

2. 定义Query方法获取GET请求对应的参数，POSTFORM方法获取POST请求参数

   query方法：调用请求体里的URL(表示被请求的URL)的Query方法，Query方法解析请求参数并返回匹配的values类型的变量，values定义为

   ```go
   type Values map[string][]string
   ```

   values用于查询请求参数以及构建对应的值。

   获取values变量后，调用values的Get方法，该方法获取该键对应的字符串数组里面的第一个值

    

   PostForm方法：

   调用请求体里的FormValue方法

   ![image-20221025214842875](D:\typora\笔记\photo\image-20221025214842875.png) 

   ![image-20221025214901728](D:\typora\笔记\photo\image-20221025214901728.png) 

   两者实现方式不同的原因：GET获取的是URL中的参数，因此要获取到请求体中的URL，调用URL对应的方法来获取键对应的值。而POSTFORM获取的是表单数据，因此调用的是请求体对应的查询表单的方法

3. 定义一个设置上下文结构体里状态码的方法，不仅要设置Context里面的状态码，**还要改变响应里面的状态码**(为什么改变的是响应里面的而不是请求的，因为状态码是表示响应的结果，是针对于响应状态而存在的)。调用上下文里的Writer(ResponseWriter)的`WriteHeader`方法，该方法发送HTTP回复的头域和状态码

4. 定义设置HTTP头域键值对的方法(今天主要用于设置"Content-Type"的值)

5. 定义以字符串方式生成HTTP请求响应的方法，涉及三个步骤：

   (1) 设置HTTP头域(设置"Content-Type")

   (2) 设置状态码

   (3) 调用上下文的Writer字段的Write方法，向连接中写入作为HTTP的一部分回复的数据

   **正确的调用顺序应该是Header().Set 然后WriteHeader() 最后是Write()**，可以这样子理解，要先通过Set方法设置头域键值对，然后通过WriteHeader方法把回复的头域和状态码发送出去，最后在响应体里写入部分数据，若先调用WriteHeader，即先发送了，再设置，也不会影响对应的响应体，即不会生效

6. 定义以JSON方式生成HTTP请求响应的方法，前两个步骤与第五点前两步一样，要想输出json对象，就要将json对象写入输出流，而json包里有一个将json对象写入输出流的结构体`Encoder`，需要使用NewEncoder方法创建一个Encoder结构体，传入参数为要写入json对象输出流，即上下文的响应体，最后调用该结构体的`Encode`方法，传入参数v，该方法将v的json编码写入输出流，并会写入一个换行符

7. 定义给响应体写入数据的方法，这里只需要设置状态码即可，然后调用上下文里的响应体的Write方法

8. 定义以HTML方式生成请求响应的方法，首先设置HTTP头域和状态码，然后调用调用上下文里的响应体的Write方法，往里面写入html数据(一般是string类型)

9. 把路由单独提取出来，并把`HandlerFunc`的类型从`func(http.ResponseWriter, *http.Request)`改为 `func(*Context)`，里面不仅包含了请求体和响应体，还包含了请求方法和路由路径以及状态响应码，方便快速访问这些常用的属性，并把handle方法参数改为上下文。

10. 微调框架入口。

    (1) 结构体`Engine`包含字段为值为上下文的键值对

    (2) ServeHTTP方法由直接在路由表中根据请求路由查找处理方法改为根据响应体和请求体创建一个新的上下文实例，然后调用engine结构体里的路由器结构体的handle方法，把新创建的上下文实例当作参数传进去，在handle方法里根据请求路径调用对应的处理方法



###### day3(难点)

照着打代码都能打错。。。。。

最后还是先把错误版本提交到github，然后再把正确代码一个个复制过来，通过版本控制来比较哪里不一样从而找出错误

![image-20221026011504668](D:\typora\笔记\photo\image-20221026011504668.png) 

  

今天实现了通过trie树(前缀树)是实现动态路由添加以及匹配 

1. 设计树节点的结构体，结构体包含四个四段

   1. pattern，用来表示待匹配的完整路由，例如：/hello/:name，即GET方法的第一个参数
   2. part，用来表示路由中的一部分，例如hello，:name，即完整路由根据符号 / 分隔后一段字符串，后面用来作为树节点的值
   3. children，子节点，用来存储每个节点的孩子节点，例如有两个路由，/hello/first，/hello/:name，最终在树上就会展现出，hello的children包含了两个节点，它们的值分别是first、:name
   4. isWild 布尔值，用来表示当前的节点的值是否含有 : 或 * ，这两个就是用来匹配动态路由的，若包含两者中的一个，则isWild为true

2. 编写在trie树中插入节点的方法(深度优先搜索的思路)，即通过GET方法把第一个参数添加到前缀树中，具体实现如下(调用addRoute函数)：

   1. 先把路由路径根据 / 分成几段字符串，存在变量parts里面，每一段代表一个节点的值

   2. 根据请求方法(get/post/delete等)，以请求方法为键查找路由是否存在该方法对应的根节点，若不存在则新建一个节点作为该方法对应的头节点

   3. 以查找到的头节点，调用insert方法，传入申请的完整路径、parts以及从0开始的树高(同时可以作为下标索引)

   4. insert方法首先判断树高和parts的长度是否一样，若一样则说明parts中的字段值都被添加到不同节点了，此时给调用该方法的节点(同时也是叶子节点)的pattern赋值为完整的请求路径，用来在后面查找匹配的时候表示是否匹配成功，若树高达到了要求但pattern为空说明该请求的路径在trie树中并没有对应的路径，只有树高达到要求且节点的pattern不为空，才说明在trie树中添加过和请求路径匹配的动态路径。即：`/p/:lang/doc`只有在第三层节点，即`doc`节点，`pattern`才会设置为`/p/:lang/doc`。`p`和`:lang`节点的`pattern`属性皆为空。因此，当匹配结束时，我们可以使用`n.pattern == ""`来判断路由规则是否匹配成功。例如，`/p/python`虽能成功匹配到`:lang`，但`:lang`的`pattern`值为空，因此匹配失败。

      若上述判断不为真，则不返回，根据树高的值来取parts中的值(姑且记为part)(由于树高从0开始，每次调用insert就加1，相当于就是个下标索引)，并调用当前节点的matchChild方法，传入part

      ```go
      // 第一个匹配成功的节点，用于插入
      func (n *node) matchChild(part string) *node {
      	for _, child := range n.children {
      		if child.part == part || child.isWild {
      			return child
      		}
      	}
      	return nil
      }
      ```

      遍历节点的所有孩子节点，若存在一个孩子节点的值等于part或者说孩子节点是以 / 或者 * 开头(说明是用来匹配动态路由的节点)，此时说明在该trie树中已经有相同的前缀了，返回该孩子节点，否则遍历完，返回空

      回到insert方法里面，判断刚刚的返回值，若为空，说明树种没有对应的节点，新建一个节点，节点的part字段赋值为上面的part，isWild根据part的开头是否包含 / 或者 * 来赋值真或者假，其他字段为默认值，并把该节点添加为调用该方法的孩子节点

      最后调用上述得到的孩子节点的insert方法，直到树高等于parts数组的长度，此时说明所有节点已经插入树中，并给最后一个节点的pattern赋值为该请求的完整路由路径

   5. 最后给router的字段handler里面添加上路径对应的处理方法，这里的路径是 请求方法 + “-”  + 完整路由路径

3. 编写在trie树中匹配动态路由的方法

   1. 先用变量sPath把请求的路由解析(以 / 把路由分成多段字符串)后的结果存起来，定义一个哈希表`params`来存储动态路由(trie树种的路由)对应的请求路由里的变量情况，然后根据请求方法查找是否有对应于请求方法的根节点，若没有，则直接返回nil，nil

   2. 调用根节点的search方法，把sPath传进去，同时也把0，作为高度的参数传进去

   3. 首先判断树高和sPath的长度是否相等，相等则说明此时对于请求路由的路径匹配完毕，或者判断调用search方法的节点是否以 * 开头，若是以 * 开头，也满足情况，则执行步骤4，否则执行步骤5

   4. 判断该节点的pattern是否为空，若不为空则说明在trie树中有匹配的动态或者静态路由，对于 ：匹配来说就是sPath的长度和树高相同，对于 * 来说则不是相同，因为在路由插入的时候，在路由解析成多段字符串中有一个这样的操作

      ![image-20221027191159530](D:\typora\笔记\photo\image-20221027191159530.png) 

      则是若字段以 * 开头，则说明分段到此就行，后面的内容不会起到作用，可有可无，而在插入的方法中

      ![image-20221027191334034](D:\typora\笔记\photo\image-20221027191334034.png) 

      由于这里的parts只是包括到含有 * 字符的字符串，因此也会满足长度等于高度，即节点值以 * 开头的节点的pattern字段会被赋值为完整请求路径，该节点被视为最后一个节点

      因此sParts长度等于树高且节点的pattern不为空或者节点值以 * 开头会返回调用该方法的节点为结束，而若pattern为空说明trie树其实并没有注册相同长度的路由，因此返回空

      执行完步骤4后执行步骤6

   5. 以height为下标索引取出sParts中对应的值，并把此作为参数调用该节点的matchChildren方法

       ```go
       func (n *node) matchChildren(part string) []*node {
       	nodes := make([]*node, 0)
       	for _, child := range n.children {
       		if child.part == part || child.isWild {
       			nodes = append(nodes, child)
       		}
       	}
       	return nodes
       }
       ```

      遍历调用该方法的孩子节点，判断该孩子节点的值是否和传入进来的字段值一样，若一样则说明找到相同前缀，则把其加入到`nodes数组`中(用来保存成功匹配的节点)，或者该节点的isWild字段是否为true，因为若为true，则说明该节点是动态匹配请求路由的节点，也相当于是有相同前缀，因此也加入到nodes数组中。最后返回nodes数组，回到search中

      search获取到nodes数组中，遍历该数组，并且调用里面的值的search方法，依旧传入sParts数组，传入的高度加1。即相当于进入下一层，不断深搜，不断调用孩子节点的search方法，直到进入符合进入步骤4的条件。上一层的search接收到下一层的search传来的节点后，先判断是否为空，若为空，则说明该trie路径并不匹配请求的路径，继续遍历nodes数组，若不为空则继续返回节点到上一层search，直到最初的search

   6. search返回的是注册到trie树中的完整路径对应的那个节点，例如：/hello/:name 对应的节点。步骤2调用完search后会获取到一个节点，若该节点为空，则直接返回nil，nil，说明匹配失败。否则，获取到节点的字段pattern则是完整的注册路径，然后对该注册路径，以 / 为分隔符进行解析，分层多段字符串，并且对解析结果进行遍历，由于注册的路由解析后和请求的路由解析后长度一样，所以遍历解析的注册路由相当于遍历解析的请求路由。若其中一段字符串以 ：开头，则说明该字符串是用来匹配动态路由的，此时去掉 ：字符，并且以这个注册路由中的该字符串为键，请求路由中对应的字符串为值，添加到步骤1定义的哈希表`params`中，然后继续遍历解析后的注册路由。若字符串以 * 开头，依旧是以去掉 * 字符的该字符串为键，以从遍历位置开始(包括)，把剩余的解析后的请求路由中的字符串全部用 / 连接起来作为值添加到`params`中，然后退出循环。循环结束后把search获取到的节点以及params返回到handle中(一开始是通过handle方法调用匹配动态路由的方法)

   7. 判断步骤6中获取的节点是否为空，若不为空则把步骤6中接收到的params赋值给Context新加的params字段，**该字段就是提供对路由参数的访问**，并且把请求的方法和请求的路径通过 - 连接到一起，并以此为键去router中寻找对应处理方法，并把传入的上下文传入该处理方法中以此来调用该处理方法。若节点为空则向浏览器中输出一条信息表示路径未匹配到。

4. 在router结构体中添加一个字段`roots`来存储每种请求方式的Trie 树根节点

5. 对Context对象增加一个`字段Params`，来提供对路由参数的访问，

   增加一个`方法Param`，来提供通过动态路由中的参数来获取匹配到的请求路由中的值(和gin一样)



###### day4

今天实现了路由的分组控制，由于路由分组大部分情况都是以相同前缀来区分，因此相当于在分组时把对应的分组前缀添加到前缀树中，在该组下把路由绑定到trie树中时只需要先在trie树中找到值为该组对应值(例如：v1)的节点，然后再在该节点下进行相应的路由添加即可。

1. 先定义一个表示组别的结构体RouterGroup，里面包含四个字段值

   1. prefix：表示在该分组下的路由对应的前缀

   2. middlewares：类型为HandlerFunc，即func(c *Context)，用来保存应用在该分组上的中间件

   3. parents：用来保存当前分组的父亲是谁，用来支持分组嵌套，即知道该分组是处于哪个父分组下

   4. engine：为了让该结构体对应的实例具有访问Router的能力，从而能够进行路由的规则的映射，即想要如下调用GET方法进行路由规则映射

2. 修改结构体Engine，此时相当于Engine是最顶层的分组，因此在Engine中增加两个字段

   1. 匿名字段 Router：一是为了表示它也是属于分组(最顶层的分组)，二是为了保证Engine结构体的实例有RouterGroup所有的能力
   2. groups：类型为`[]*RouterGroup`，作为顶层分组，应该保存整个路由中所有的分组情况

3. 修改 gee.go 文件中的部分代码，将和路由有关的函数，交给RouterGroup实现

   1. 修改Engine构造器：

      ```go
      func New() *Engine {
      	engine := &Engine{router: newRouter()}
      	engine.RouterGroup = &RouterGroup{engine: engine}
      	engine.groups = []*RouterGroup{engine.RouterGroup}
      	return engine
      }
      ```

      先通过router的构造器生成一个router，由此来得到一个engine实例。由于engine是顶层分组，因此它的前缀字符串和父分组都直接采用默认值(为空)，而里面的engine字段则是自身，同时，它还需要记录整个engine的所有分组，**再次强调由于engine是顶层分组，因此它自身也属于一个分组**，因此记录分组情况的数组里面要把自己添加进去

   2. 增添Goup方法，用于构造分组

      ```go
      func (group *RouterGroup) Group(prefix string) *RouterGroup {
      	engine := group.engine
      	newGroup := &RouterGroup{
      		prefix: group.prefix + prefix,
      		parent: group,
      		engine: engine,
      	}
      	engine.groups = append(engine.groups, newGroup)
      	return newGroup
      }
      ```

      里面需要传入一个参数，即分组的类别(字符串)，**所有分组都共享同一个engine实例，即最开始通过New函数生成的engine实例**，因此新分组的engine实例用的是调用该方法的group的engine实例，这样子就能保证共享同一个engine实例。新生成的分组的前缀为调用该方法的group的前缀加上传进来的参数，这样子形成嵌套关系，且它的父分组即为group，最后还要把新生成的分组添加到engine.groups下，保证记录了所有分组

   3. 修改addRoute方法

      ```go
      func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
      	pattern := group.prefix + comp
      	log.Printf("Route %4s - %s", method, pattern)
      	group.engine.router.addRoute(method, pattern, handler)
      }
      ```

      若使用engine来调用addRoute，相当于在顶层分组上直接添加路由到trie树中，与实际情况不符合，因此要通过RouterGroup来使用，得到的最终路径为调用该方法的RouterGroup的前缀加上传进来的路由路径，这样子才能体现该路由处于分组下。由于得到的是最终路径，即从trie树头节点开始的路径，因此通过RouterGroup中的engine来调用router中的addRoute

   4. 分别修改GET和POST方法

      ```go
      func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
      	group.addRoute("GET", pattern, handler)
      }
      
      // POST defines the method to add POST request
      func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
      	group.addRoute("POST", pattern, handler)
      }
      ```

      原因和第三点一样，若是直接通过engine调用，则无法在trie树中实现真正的分组，直接调用相当于在trie树中直接添加传进来的路由，通过RouterGroup调用，**才能把该分组对应的前缀也传进addRoute方法中，从而得到正确的完整路由路径(从trie树头节点开始的路由路径)**，例如：

      ```go
      r := gee.New()
      v1 := r.Group("/v1")
      v1.GET("/", func(c *gee.Context) {
      	c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
      })
      ```

      此时通过分组v1来进行路由规则映射，若直接调用engine.GET，则最终映射在trie树中的路由路径是`/` ，而不是 `/v1/` ,因此分组的时候只有通过RouterGroup来调用，才能在trie树中正确的映射路由

   5. 可以仔细观察下`addRoute`函数，调用了`group.engine.router.addRoute`来实现了路由的映射。由于`Engine`从某种意义上继承了`RouterGroup`的所有属性和方法，因为 (*Engine).engine 是指向自己的。这样实现，我们既可以像原来一样添加路由(即通过Engine.GET直接添加路由到trie树下，因为Engine所属分组前缀为空，相当于直接在trie树下添加路由)，也可以通过分组添加路由(通过RouterGroup.GET来添加，此时就会先把RouterGroup的前缀和所传参数进行结合，然后再把结合后的结果当作路由添加到trie树中，保证了添加的路由处于某一分组下)。

4.  



###### day5

今天在框架中实现了中间件机制，即如何把中间件应用到对应的组中，同时搞定了如何通过Next()方法实现中间件和处理方法的逐步调用。

Gee中间件的定义与路由映射的Handle一致，都是`func(c *Context)`，**插入点是在框架接收到请求并且初始化Context对象后**，这点可以在ServeHTTP中体现出来。

Gee的中间件支持用户在请求被处理前后(即处理方法被调用前和调用完后)，做一些额外的操作

`c.Next()`表示等待执行其他的中间件或用户的Handler

1. 给`Context`结构体添加两个字段

   1. handlers：类型是[]HandlerFunc，作用是记录该请求要执行的所有中间件以及处理方法
   2. index ：int类型，作用是指代上面的handlers中的中间件或者处理方法执行到第几个了

2. 修改以下newContext函数(即Context的构造器)，在返回的`*Cntext`里面添加一个字段 index: -1。把index字段初始化为-1，则第一次调用Next方法的时候index会加一，刚好执行handlers中的第一个中间件/处理方法

3. 添加Context绑定的方法Next。

   ```go
   func (c *Context) Next() {
   	c.index++
   	s := len(c.handlers)
   	for ; c.index < s; c.index++ {
   		c.handlers[c.index](c)
   	}
   }
   ```

   index++是为了改变索引，使index指向准备执行的第一个中间件/处理方法，然后逐个调用handlers中保存的中间件/处理器。

   `index`是记录当前执行到第几个中间件，当在中间件中调用`Next`方法时，控制权交给了下一个中间件，直到调用到最后一个中间件，然后再从后往前，调用每个中间件在`Next`方法之后定义的部分。

4.  例子：

   ```go
   func A(c *Context) {
       part1
       c.Next()
       part2
   }
   func B(c *Context) {
       part3
       c.Next()
       part4
   }
   ```

   假设我们应用了中间件 A 和 B，和路由映射的 Handler。`c.handlers`是这样的[A, B, Handler]，`c.index`初始化为-1。调用`c.Next()`，接下来的流程是这样的：

   - c.index++，c.index 变为 0
   - 0 < 3，调用 c.handlers[0]，即 A
   - 执行 part1，调用 c.Next()
   - c.index++，c.index 变为 1
   - 1 < 3，调用 c.handlers[1]，即 B
   - 执行 part3，调用 c.Next()
   - c.index++，c.index 变为 2
   - 2 < 3，调用 c.handlers[2]，即Handler
   - Handler 调用完毕，返回到 B 中的 part4，执行 part4
   - part4 执行完毕，返回到 A 中的 part2，执行 part2
   - part2 执行完毕，结束。

   ![image-20221029114633338](D:\typora\笔记\photo\image-20221029114633338.png) 

   当index<len时，会进入for循环，因此会执行下一个中间件，而当返回时，index>=len时，会直接退出循环，然后回到调用该Next方法的中间件里执行剩下的代码

5.  定义Use函数

   ```go
   func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
   	group.middlewares = append(group.middlewares, middlewares...)
   }
   ```

   这部分所做的操作就是把传进来的中间件依次添加到该分组的middlewares字段中，该字段用来保存作用在该分组上的中间件

6.  修改ServeHTTP方法

   ```go
   func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
   	var middlewares []HandlerFunc
   	for _, group := range engine.groups {
   		if strings.HasPrefix(req.URL.Path, group.prefix) {
   			middlewares = append(middlewares, group.middlewares...)
   		}
   	}
   	c := newContext(w, req)
   	c.handlers = middlewares
   	engine.router.handle(c)
   }
   ```

   if判断的是请求的路劲如果的包含某个分组的前缀，则说明该请求属于那个分组之下，因此在最终执行处理函数前肯定要执行该分组下的中间件，因此定义一个HandlerFunc切片来存储要执行的所有中间件，跳出循环后用传入的参数初始化生成一个Context，并且把保存的中间件赋值给c.handlers，这样子才真正的做到在处理前执行中间件(因为处理函数处理的对象是*Context实例，也就是这里的c)

   

7.  修改handle函数

   ```go
   func (r *router) handle(c *Context) {
   	n, params := r.getRoute(c.Method, c.Path)
   
   	if n != nil {
   		key := c.Method + "-" + n.pattern
   		c.Params = params
   		c.handlers = append(c.handlers, r.handlers[key])
   	} else {
   		c.handlers = append(c.handlers, func(c *Context) {
   			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
   		})
   	}
   	c.Next()
   }
   ```

   因为上面的ServeHTTP方法只是添加了要执行的中间件，并未添加处理函数，因此在执行路由匹配后还需要把完整请求路径对应的处理函数添加到c.handlers中，最后调用Next方法开始执行中间件和处理函数

8.  



###### day6

今天实现了HTML模板的渲染以及静态资源服务

1. 定义一个绑定RouterGroup结构体的用来创建静态处理器的方法。绑定RouterGroup而不是engine是因为有时候进行模板渲染是通过细分的组来进行的

   ```go
   func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
   	absolutePath := path.Join(group.prefix, relativePath)
   	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
   	return func(c *Context) {
   		file := c.Param("filepath")
   		// Check if file exists and/or if we have permission to access it
   		if _, err := fs.Open(file); err != nil {
   			c.Status(http.StatusNotFound)
   			return
   		}
   
   		fileServer.ServeHTTP(c.Writer, c.Req)
   	}
   }
   ```

   `path.Join`：将任意数量的路径元素放入一个单一路径里，会根据需要添加斜杠

   <img src="D:\typora\笔记\photo\image-20221030162303647.png" alt="image-20221030162303647" style="zoom: 50%;" /> 

    

   `http.FileServer(root FileSystem)`：返回一个使用FileSystem接口root **提供文件访问服务**的HTTP处理器。

   FileSystem定义如下：

   ```go
   type FileSystem interface {
       Open(name string) (File, error)
   }
   ```

   该接口实现了对一系列命名文件的访问。文件路径的分隔符为 /

   `StripPrefix`：返回一个处理器，该处理器会将请求的URL.Path字段中给定前缀prefix去除后再交由h(第二个参数)处理。StripPrefix会向URL.Path字段中没有给定前缀的请求回复404 page not found。

   最终会返回一个方法，该方法会用来处理静态文件

2. 定义注册处理静态文件的方法

   ```go
   func (group *RouterGroup) Static(relativePath string, root string) {
   	handler := group.createStaticHandler(relativePath, http.Dir(root))
   	urlPattern := path.Join(relativePath, "/*filepath")
   	// Register GET handlers
   	group.GET(urlPattern, handler)
   }
   ```

   Dir定义：

   ```go
   type Dir string
   ```

   Dir使用 限制到指定目录树的本地文件系统 实现了http.FileSystem接口。空Dir被视为"."，即代表当前目录。

   该方法注册静态文件的处理方法

3. 给Engine结构体增加两个字段

   1. `htmlTemplates`：类型`*template.Template`，代表一个解析好的模板
   2. `funcMap`：类型`template.FuncMap`，定义了函数名 字符串到**函数**的映射，每个函数都必须有1到2个返回值，如果2个则后面一个必须是error接口类型。如果有两个返回值的方法 返回的error非nil，模板会执行中断并返回给调用者该错误

   前者将所有的模板加载进内存，后者是所有的自定义模板的渲染函数。

4. 设置funcMap字段

   ```go
   func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
   	engine.funcMap = funcMap
   }
   ```

   从外部传入一个对应类型的值来修改engine中的funcMap字段

5. 加载html模板

   ```go
   func (engine *Engine) LoadHTMLGlob(pattern string) {
   	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
   }
   ```

   `New(name string) *Template`：创建一个名为name的模板

   `Func(funcMap FuncMap) *Template`：向调用该方法的模板的函数字典里加入参数funcMap内的键值对。若funcMap某个键值对不是函数类型或者返回值不符合要求会panic。但是，可以对调用该方法的模板的函数列表的成员进行重写。方法返回该模板以便进行链式调用。

   `ParseGlob(pattern string)(*Template, eror)`：解析匹配pattern的文件里的模板定义并将解析结果与调用该方法的模板t关联。如果发生错误，会停止解析并返回nil，否则返回(t, nil)。至少要存在一个匹配的文件。简单来说就是传入文件的路径，该方法对路径内的模板定义进行解析。

   `Must(t *Template, err error) *Template`：用于包装返回(*Template, error)的函数/方法调用，它会在err非nil时panic，一般用于变量初始化

6. 在 结构体`Context` 中添加了成员变量 `engine *Engine`，这样就能够通过 Context 访问 Engine 中的 HTML 模板。实例化 Context 时，还需要给 `c.engine` 赋值。

7. 对原来的 `(*Context).HTML()`方法做了些小修改，使之支持根据模板文件名选择模板进行渲染。

   ```go
   func (c *Context) HTML(code int, name string, data interface{}) {
   	c.SetHeader("Content-Type", "text/html")
   	c.Status(code)
   	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
   		c.Fail(500, err.Error())
   	}
   }
   ```

   `Excute(wr io.Writer, data interface{}) (err error)`：将解析好的模板应用到data上，并将输出写入wr。如果执行时出现错误，会停止执行，但有可能已经写入wr 部分数据。模板可以安全的并发执行。 

   `func (t *Template)ExcuteTemplate(wr io.Writer, name string, data interface{}) (err error)`：类似Execute，但是使用名为name的 t关联的模板产生输出。

   





###### day7

今天是实现了错误处理机制，即定义了一个能够捕获错误的中间件，在框架中使用该中间件能够捕获错误，防止程序崩溃，保证程序正常运行

1. 实现中间件Recovery：

   ```go
   func Recovery() HandlerFunc {
   	return func(c *Context) {
   		defer func() {
   			if err := recover(); err != nil {
   				message := fmt.Sprintf("%s", err)
   				log.Printf("%s\n\n", trace(message))
   				c.Fail(http.StatusInternalServerError, "Internal Server Error")
   			}
   		}()
   
   		c.Next()
   	}
   }
   ```

   实现过程比较简单，就不详细叙述了，因为Recovery是一个中间件，因此肯定也需要调用Next()方法来执行后面的中间件和处理方法

2.  定义一个trace函数，用来获取触发panic的堆栈信息

   ```go
   func trace(message string) string {
   	var pcs [32]uintptr
   	n := runtime.Callers(3, pcs[:]) // skip first 3 caller
   
   	var str strings.Builder
   	str.WriteString(message + "\nTraceback:")
   	for _, pc := range pcs[:n] {
   		fn := runtime.FuncForPC(pc)
   		file, line := fn.FileLine(pc)
   		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
   	}
   	return str.String()
   }
   ```

   `uintptr`：能存储任何类型的指针类型，可以进行指针计算

   `runtime.Callers`：函数把当前go程调用栈上的调用栈标识符填入切片pc中，返回写入到pc中的项数。实参skip为开始在pc中 记录之前所要跳过的栈帧数，0表示Callers自身的调用栈，1表示Callers所在的调用栈。返回写入p的项数。可以理解成把指向相应栈信息的指针逐个放进切片中

   `str.WriteString`：strings.Builder通过使用一个内部的slice来存储数据片段。当开发者调用写入方法时，数据实际上是被追加到其内部的slice上。WriteString方法可以理解成往slice中写入数据。

   `runtime.FuncForPC`：返回一个表示调用栈标识符pc 对应的调用栈的*Func；如果该调用栈标识符没有对应的调用栈，函数会返回nil。**每一个调用栈必然是对应某个函数的调用。**

   `FileLine`：返回该调用栈所调用的函数的源代码文件名和行号。如果调用栈标识符pc不是f内的调用栈标识符，结果是不精确的。

   

   具体解释是：

   在 *trace()* 中，调用了 `runtime.Callers(3, pcs[:])`，Callers 用来返回调用栈的程序计数器, 第 0 个 Caller 是 Callers 本身，第 1 个是上一层 trace，第 2 个是再上一层的 `defer func`。因此，为了日志简洁一点，我们跳过了前 3 个 Caller。

   接下来，通过 `runtime.FuncForPC(pc)` 获取对应的函数，再通过 `fn.FileLine(pc)` 获取到调用该函数的文件名和行号，打印在日志中。

   

3. 











​     

​     
