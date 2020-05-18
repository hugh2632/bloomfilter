package bloomfilter

import (
	"github.com/go-redis/redis"
	"github.com/spaolacci/murmur3"
	"hash"
	"hash/crc64"
	"hash/fnv"
	"runtime"
	"testing"
)

var testData = []string{
	"二七N9P2六LOPL97TOPE三六十X3cCOC6c九M8八7TV4T十3GJMJ七bMIGTFBXD八I六二L七cF85S96七EDB十EWHbW八aUcJV8七aWXXGKVHLK四L1KeF",
	"E5九6SKVb9四BOa7XP5H二8十4L三bbTZHOG1FB八NAX一FTATMa四Y七MZYLd4eAAN六五G三AI0V八YTRKZCUTJc65O4M五五三71一b6一E四0O3R42O",
	"O十S一L六ZZKI一六P五Sb2GG4ENC4Od55CI0JBB79A九29三XC二十KaXRCS33八Ve4ZENOWIFO一F四0PJ一七cKHE29KOL五Dd六VXS628BVI3V六XB",
	"F五4九dMB二M4WNe6四FBXX0RQJSSKDK19JE一D六1ZA8U64二IIWUK5Zd6VAb八八G一eHT七八J1KDdeUXZTO9WMZb0XadacMJPCaUVGS七F1一三",
	"8C5五KZ6aUM0三AMWQGEb七KI十cUYbWSUA1S二Q97十十0GGa0XCXL3EKVbP1八一T六IKA二6十beS五二RR四一eWEKH四VF5LbcON七9七八V六YICN三四",
	"十六5bWC86十WZ46SRTd0cQ4五ZIc九TDHTI6TaWC七OFHCde56五75Y十8H十二8CVHP7V四9九六WXU2S二A8RbOd七6十Ja六aW九0MK七c7Y十W1GPeY",
	"GJCA七bG9bVTOYb2deIASACDKW525FW七5RA九DEXPE一FK三MWBZA2d0V十N九X9三KMZ二Z25S4ZTFaLDGZ四E6RR九ePX9B8aZ五Ce7T七I二Ae",
	"4PCVcMMTJRN072九aP八d7T四dE90UM9T7BYY四九P三ODK1六X六PG02Z十7ABVCZRLTB一TG5一一U79一e一WKICBGOL9bMVa九I9F7bXbE4IY6三",
	"1TZIKRW九ZTb四R0XM5三MS4二ZCXV六ReBUKWJP7四1de十三三4ASJE9a4MFCW八YSAGVca二VA八十CBPRCXGYbUcTEANU十KL1四TeJL四三QT十P三",
	"五dW一九X二7O二WJI五八VLY一MCUHdO2GWdb十TXLA九A五JSVXb6OW6K十2HBDGZNT十八LP8五BU六五UKRb22一七Rd1W0六B5一RY六I5十YHcO8八EJ五2",
	"Q4c一0OI5三bSbPdTMNTe3HBOJ五六2UW3X6八二S2七SUEECA7BB五四NVOdSWeWL八9十六六0CG三C1ZPWIb2VeX八7PL一Q6四Z六FcFd六NEP4GD0L",
	"Z49Z四六XMVLHJID1IS8H5F二五F八AS五九WFYG3G1九QQd9N六VHTLJW三A6ZEb2OUGO三1PJ二6XV1N四a6WEA九bUNYFN3六2QD四XDB8KGQV八DH",
	"八JE八IKWX四九F八九八3U一W2IJF1cVYL三P六三O四FeCP十UcM一SLGENPV5DPSS九Re49LN2L六一e3三五十KBeY2六9BA十e4b二2ZNCe七3Z一Wa2XbXM",
	"0AbLCXZc八一4HB7OL7eRY0HOW5EV九VRH36SSNQW十5九5F九Qe六0二RN1S21VAAY0aQV5S0四X6c8JHLcae2d9X3一0I六LLHDcHHIKLV4K一",
	"M二9QXWY十四GWE五D五GU0VWWGVD一N3S0V759YKJ二RTYdca5E四T一O一LdHWdOd七XPL四MY一8DDJLP6十82G八R五WT2OEBOKTJa78b一ENPX七d",
	"BNQSTZ八M三十RGY78V二U5Be561O四三三九OZSI7EPEJ九KMRH0S6EW四NRRIA93I一8三3ASHYLV四一OWRMa1J3e49八RJ一0a7九一J七七a3Oe八TH4",
	"JH十六ZbP0CXWLa6B022TZ2a0八E6IbAd6一CFTFJ8十Z1四十6六EEaWTJSFc1L0C2acN2NL7Qe五dA1六WJX五A70I6dU2c1C五NN二一3IJ23S十",
	"Rb2V十KUSM3e四DM五4T3e7McUV7dTaMb三5WZHB4W九RT3KXEBFLT8YD2十1OKMDO47YIb八FaOdJU四六DY09M3R八T八MR十KIMNdAN5DA4八N",
	"ETY4O三CP2JM6S十HB十EV五JX五3L6三六WUBL7Id十I92四U6U七Y6X十5三十a六Z四QD八GJGD四ES2222PAU七NOQWQBc七IU2UKH8Ud7A3F一GA8Y9",
	"一ACJ10九bGUB54十C6FUKK7b十十五R5Bd六六d五六5I六八c四1XMRECYV九2EXL0AbZaXU9aA五五VM九cR九NC三1八XV三A三6THL3d59九cFRec九U36O",
	"81M1二3P五十c一Y三BbA04九W2AbLU九dTOGSEYU十三一d九CK50O四WNYTAKAD8五WNGWLG10KHL六X一HJNEMSLI二aGI1MR十6O一C二G十WD六8QRYF",
	"F15BRSTU2GWVA8I9TQG3KEE7R七LQV8KGe4bdCe1一ba十M6a02dK三一LI2IK85二九8JTX1RKDN55e八COOFYbL39LXZX6V一7dK五A5一cBD",
	"6DJFP5X一6一5N七IRVQ七X七27Y三N6BG83eZ一Y八1DFMAE94七四O四Z3U1Y六eS三QUWPH3二dJBE二LPbMMKS十X8HD1七cJ九七IeK九S4七二T九94一d",
	"四JCaX6SQ四V9bEDa3三五Z七ILI7二一五Ib1LbFL五JAcW3四6aHRXa4K七RWW5b6d4三MN八GI六N一6八AUH0D0EBVIVdF三1十J二XYSV七HN三X四二aV",
	"M六三九dcJT5三aE一3A0XM二HXcd七四RXcW0九96四SeQ八eP六三HQ十H9五JH1J3十九eOQF六cYE六6J四W4O二843ICF74DUVXXDCV九8一a八八Z五6二KIc",
	"cUWX1WX50六1八US七CG一VEE六86五2G四DIUbVX1七DOF5b80941VR7二GV2三CY八二十DD十J二35LTKXU2Lc五bedaJ3L1九ZC五7eZHbeZ一W三五S5",
	"A九九HZ80PFW七F57SEcC2S九F七FPFMaDAE八1e1Y6PU二ATASYIHQSMcLK7八bW8五T九6803十e六TNP二SC4X0JCYF8RH6UE五OAF0WKWMJP四C",
	"八RSbGP七Q六QMR9KW4四九B八RFT6DY二YaAeBP一D4五17U3六WRN4六e一a二JA五0PNT1PMZBa4YP七九Teb一T3dcRJ七X3aM九GB一SU八S3DU一ET五2",
	"YK二八BKHZP56O1V四5dB七J2EB3一27Qa八dP一06H0HE0十5NX7dH七十XLK三0CZFZQ五Q47K2OIEWBV七6EA五L2EY5KGPAJFc2A四ELRNOOA2b",
	"CcbQSJKWd69X3十YeWQA0d九四YHF一EW8六81I54H0二六五D6十SS六9eV十九DDb946九Q三5GLBAHVZ6八dXaW三G3一S43U81ONXIScPR七E2c二8七",
	"6dYA7I四a8GIaERUcJD900JT九CY四P七d八H4三V1三一C5八EA4V八FOHOILQ9LTRFBF5五YA四六G九I0QC3PE3dS9NQB十J九a七8BACJ3四七XG五L六",
	"U一HdORX四KZ74BJcC一VUK五八B二0三GOd2XS8三VX十Bc3CdUH844AO一B1十F3V23G17HT八九Y3MQA7九d九KLBLVa4一十QR8S4A九00J八Q二R6二d",
	"VAdCMIS九HVI96B0dGJKL7P四5三H一4四LAX0ad九七5BC九E5FJTGQC五六cXN七0A9LdN十三G3aB9B70UcN3六RG2二TSQ537NM三47五DFd九EZ4T",
	"二O四c2cYWN1二5GO四6e五b27十MD7四Wbb二cVbFLbYcc一Y1JTIU一3七E20ZGcI5E十GLJS五09五8dV7RG三四T2d九六2EC五G8一QFQ四JdI7RTSDG",
	"O7N三二KQ59IH5K8R2LQT一b一T4G五1V8AQ六UMXKQTe八九C十V0VB22T8PEJH八YN五3U十4aB2a一5五二四6六BB7七十FL6Md四3CJ三AM5YS五NKeUa",
	"dA九7JKJ0十V六7ZJbb9DaW七cKZ四N七H四五Fd十O六二bc9六AFU7七J一C7O6DKBII一8十EUR八EOdDXA9七TRY5八CY九d六2五二五TW2Y四一d1一2AUPHG",
	"E1HV八b1c四PPK五三bc611EOAcWb2七U五cP1RY九QdNe0六17一十WW八八6RW4FH七WTHX十KK09L六1cEZ4RDF五cVdY6四7aG三九8Ma四八九UXVNC6四",
	"KHe8RIYZ二QOE一O八FLRNP6TH6bZVI9二YY7FB1一403XR八3DB37一C2十b48一1I5dSI六六VWBR五三九WDCdK9八WO三四Z九四CcSL1V8PDUMBJdK",
	"KV七0XI80bRPcYFQdF2TOaWCXI2M一E0六一4HEOL118AO五L一IQU三028PX九O8A九R01二0LaZDXZPMEKH二0dQC九d7五九32V6ZUFDAPKZD九二",
	"Q八7六YVA二aU六J九NTYcP十d1QUA1七ABAFS三KX5GH0MDAL6QEZKRNd4十POa五十P1X6三J9五IWC九八RFV二Q十bTL九八二O五8X9PTHCQEbabc六八8",
	"IQbc三QQVFWJWbRW0DPaPNW五JYdLNdH七TW四20十QeEGba十X2BPEBBB4S二WacRQ2I一四D1N4cXcG0四八六一aNY4VZ三10十CcRF三KJ3TA31P",
	"四6O96FGRJc十4W9XPb9I6X44Xe6P四K四二GV4十Ga三PSbZ十Na四IOFT九U8LS四7JV7E二G九ITRG9H九十FLOZIe1三8Q二九二EMCC3P七FXAU4HSB",
	"ReZR一Fe9UFFe6SME七S十DbDP2AGMEYNL8dYYc六八cN四C6FZ3T24UR8一AGB3SUc4八UC七八JG八九六38W三五37VQ七aA9Y二PY六五RMQA9b九e二S",
	"Q八DbZ7QRa五RBMd9K五UZ四WZBUI2ZFJeI16TDMW四七V2W一GBASbW5六P二d三3DRD三ZM4二RSS3WG6UATYa7十c5二TS一69b9八B四TYL四VMHNe",
	"JKMV48bKEbeJ2aCSD三OAbJZ1HQI8bPCY十JFb五G十IIP四WC8QP2KS3EW9DSX2b6FbX2S九一9IAQIFY七7IV2D1七J2九WM三e四9GUIFcc九B",
	"六9cDEV五bIZ8aFH32S1N08五b五KN5ZdJQGQOM1八L4一XNSe二NS十二十U7FWe7ObBN3MaDJUG9ecLVUEJ21KdN7YUM1二aEG七QC0七O6三QHC",
	"ZbE8c一Y十AWZ3MVKKIK5ZP三3GJb四2R2V1AI十48SLF4EWAPN一S9dYG十78c7S8DTAH658RcN七Q2C8d6FTZ九13KNec1GYS二O336B599A",
	"H6b8LT26e7五RUB73I二Jc34九V82CRETYdBX九5EUN六aM六68eMEWRLPW3Jc634五七CH一RZGcTS93九ZUM十DYM二7QGLUI七二cF43九十UU十V3",
	"XEQd5九9GdeARaQCV二SPL7eQPKG七5c五86VOK十PP1Kb8九Z一I七BE5F二七Q97七YEWVXVZT64一CU二4TC217T九CG四八T4ce95七S9dE2六D一Z3",
	"5WcYE二c9G7FF8bAAINX48bY3一Q2一二九二TWGaea4四MTbNO6XVBU一MaGEO9RCUC十YIUIDHVOIXCOWJ三一5T2dUDH7843B三7c9e二7三cae",
	"cZU九McdFEc7十dRL十bODJN9T4六R六FK05K二六1二8Y2J3INBF二60FbF8I七0XJ十T74WAYAHD九94三ROK二SRMXa二XcK4MGTdEXc2TAVSPK七",
	"6PRU四CJdGdAbZ7三UeR1TAc6OeDWL六十五九二H3b六aBQ三eM4A三H二D7WcW180J9K八X九dPYOP六Q二9c9OE五W九九YOUFB54TIGN7N二E一RD6五d",
	"P2K五R1Z二十0K06NWQNK0TLOb六d4PXRX六12c1d1W二N1aOSAeJ五3ANM四QZdAJ4NLdNNb九四BLSE九二c551e2BUFC二65P八六7A六六F五九FWUU",
	"T4BZHOI二0Y9四bB一TRA4eQGMH七2ZCa1SYAb26八8YRQBCPPXcAX827I八R九二dJ十GUaVU3cPJMHLU3JN四8HAG1NNIJXTM三M9M九九七ZLW二",
	"RU1BULLW三e5九JX六9S3E0BDL8QDVG八GQHaVd一7K3dXEVU三aAH七七四Q八P九8A93UX4Y68M二0D三N2七5六F八八六六8Ca四b七3DKb三四PPY八YEWC",
	"44KCcDDYQ5JbJN五FT7九YXU6Pa8NXFd六UVdXNF三Ga八五IbI四UG495ZCK七daMI九B七一N7F六Z8九LAc0aT7四2YTGQ2六二K7W5五7dL十3十MZ9",
	"TFB4Va0RQ3九bY6B75T5W96九ZcFEW8OaHN8JP五3EBe89S9四5一J7I9ZXH729YKY四e七P38四aM5KeMACNbNU六七dOc49WdEGaGRG2EQH7",
	"九4十ePP九SXZPcCO3三X四N十W98PYCT七一2QDKMI七Ja0九Q六CF1M5十O七AG八VZ9B9GUVG四一ZB十十十7B二P五六十7FSQ八K1D十EO五06ROJ八5S四W十V",
	"da九322二V五R92JN7d五e二XVQI十D2EPAYX七YXHN六8HS61三384aW六AN1RYUbAVAD二G九0Y6AMUcGc十dP8EMUL74M11UJbc八76ZDKNP6VE",
	"879S2eb十bQ六6R五J四ANacQ1HHJ9六SMLFeS九4QQe6I九SSA四X39H九b7QN1一e六a9十三eZ五ZGZD五e六八六XDTX1S十2C3十e九IN8SG1IAKPUUc",
	"9二1XN四B八c五S0E四ROdcF十HTe二ANA五98W二8HRD1I二MA五GGceC02UPeDSb四R5XF十b4GT九X八HXAZQ44Q六C87cCFN2P八0九S0aJXGSD92九",
	"三KF7bYDdOb5ZOYM1TaQ5八7十Z八十F六W三七八六0KG72九Se九HDJUJDJP5四9aQ9七ce十S84BFOAMF977Ha八UZ4YHUR4BME十四六6W一NAA四G十2五",
	"aCWD7十dIWHcTeAbF4Y七Pe九Zb十七cW四8AVU七9b七VFUebM八I八4c三四ba九五X5d十KPN七eaADANEbEH五4GFWeT九bLYI0Q九十十Y4B六044c78J",
	"0FXNQZ六Bd十F六F六d一K9五SeIG11U1ZI7OFSX六Je0O七cNWSKHWSY七十dW四八ee8O九六五eN1N6V二O五三JD9M5九J七AI三TSDFRYMOQS十WNQ三4e",
	"NEQ8c5T七G十H四C八TT六GJ0771四EP6IYQPVCKBS二七RH四ADB74K6九cFGP6十四KMdJCS0RDIFdWL五1e4四2四VEX七81C58M六K二一aZS七PCEVM",
	"NRBG一QD二O60OOF九cDPT971七87DLZ9JKN六P十二EYXQOY70VcHTD九9Oe八八Q八F2十VY26ZJEaB四N7aM1MHQ九CG4四LYMYI27W一517八C4VY",
	"9WS4C7OUYN三Bb八eRFUd一TdRUYDOW四M51ZY2bV八UVE0d4HcVZ8RRE7I4L三7GVRSGHLdBJb0ET2e一9STNZ5ZObS5二十MOI88IAQ2I93",
	"bI05WCFZe一3一9十5ee六B一U一0BN五七KHZT五一bQ2204NQAN7UZaYBE4bSSGHPLDIAY0C1a八MXe六DR七NE3ZXHYRZX十四三PZ5a二2ALBHPG二",
	"FTI6C87V十89dTCW九九X一七七十a7eMN8dKFDMX0I六5YPI36F34dUR9二4LOJ四GR66U二二UQM六7XRU3WdU五C7U一NeL3Za二Fa6LETTSWMM3十",
	"C四Z四T八HSb5MI0S7NK九VE十9七dO8三JUD一HHP1Y0RHH十十3HOKJ2Z二六Vd9IPcC0TKXVK6Y5b1F7YCIaN七JUVYPaeJ1UaIR六FJJLK七PET",
	"MFS四Ca八SdU八十JLPGe五VP0Y二We二Pc八9HEEHQ四VZ773ZeZ二KMKUY1YVQ9ZF四ATPedGI78四25a89E八1BE四K四FQ一三FP四14五SVI五0九G七1",
	"五X十Ua44BVF2三ADS2五9L六QJ三五PZ7aS七aQPYX08十六O8十BVF六EJ九B二一3十T五QO一M五七TXdGX918G七Xa七MdX四三McXQe5QZB9R五23HZK61Y",
	"I四八FHTQ2九四Zb7GT5McEIP1MWGLF五26SdGG7bSMFY五WSV8F5UVSC十C一5EQ六IULE十七IRO二四c4MVBP7VEHBJ8TX4一663HeJ四QZ九1L12",
	"XAL七五7K61OLQJG十c九G九FDHVM四d八997E5c54IC3七4JZN8eNV9e一QBUSKO六XB0K302BZP二H8RIMTO八八Y一S9b4V7六XaXN八DTec一eBI七",
	"dG九一BD三TDK二9d7二HMWPbNPQNTbJ6十VN7ZKJ九aD9六十1FLY13HZAXZGZN五J8bcG2S0MMQRGCRcQZDDCMQB0N0V2IFSa80GM4aPeYYK",
	"8QNb8五Fa五四LEPRP5八B一E一QaIOcUB12JaKFK4S八AB六BVD83cbdJ03GJ五E三8Z八5XLJE9二ZS5TRXJYYZI2cBb四九AY一3六89JY七bP三A82",
	"39八9J5ONXIBJBB4E二07G九六YQGPN4H四F一一A4EebK1XG七MDY2五7MaL五三五a五九Q十Z5PL二十e6DNd9JTObYcbRSXd4PLWI一三IMT六V7九CJN",
	"4R一DERE71二H9U3cGQ六四V四dQI48AW二dNB二dLO五A1A1cAL四91TcOHO42十Bdb5SMP1KQE3NMR四2Ke一六二Z28YH0FQa3ZG六2七六Z3O五七4c",
	"a1ZU三0NOa0LYEc43三一二五17JYWPHe7AXbHZWAZJ6IH2八J3JMFKaP2E一7aM十FBe59T三SRcbLH90R9六9T7740A八IMIMO三FaQJT4cR38",
	"LdF1八L5R3Z九M四C六4EUFA4Q61GY一1九TMJLC9ZD四35W14H8ac1Hb五dWda3七bKbAHAOZ73U三FQC三G五Y六5VSaZ3九1五四FNPPEIV2BOQ九9",
	"6aBSNQRDHSKU六UVHVZ2D七0三VI2b二aO2六九四YGG六9A九八Oc2P5E7一TdM4UC十OJCCDLTDBGPBU4be八五五三8Y9DOTE二FIC二三七四E1ZX49CB",
	"七九I5P2138TBdc58b八H6LY三R六a5S九eL五JRL6六7Cb092b七J十Y5N340Le2PZ8B5dCeM五2a九Q54P二一2b2E九PO6EIPeKR十P7H六V92三e六2",
	"S五cIAK95M四D8七86二五V6YSR3E0WD一V1P61S08XKVYXcd十a九T3一b七5X七YQ二9F三eDdZ3I四一T73Ta3七OVIR三I三七0A六二三QG四CVHMU84HC",
	"WYFPS二H1WCWXZM5八KL五OTW七2四TQZbBK8X九S四L97XGOENG十22VQ1CCI五28QM四三九Za六7Y十九ee七9DJFIHMWF5NCeUKc9TEB8NcJ三RLL",
	"F一TeR6Qc二十XVKZJI八二VDXAZSZK02cRB五ea9八a42e23七一1dIZ30七cU5五Te五5Ga73九7AA4UEddNK64九VUGec27C四52GWDT五A九0U0RB",
	"QW九N2613E十9G九L七e六WYE9JaN3Hca一J八XBJeU52dWD64YA06e6十九三D五V5一deBZ六I970e十ZAZ928LT九c八4SQaeDNb九J四8六XTJZD2eC",
	"PcJcNcPA一6d七QZHXOR八bU十FKZSV8cEb29PS7cQ8R23OHO六0HU4M十a24Q八1KO五O9十EH五八O八五aDPS5OB551D5eSL一八Z0NT七0I5十三XJ",
	"MDUdHS9二CSOHW三KTY6LAJSI五N七TTJ五N五4VO6aZ9b七2MWNa40e7ZCdP5APVFAAWDZ4三3S八EWJH三L3A1R5I5KW十SPH七M4N四9R四四十S2",
	"e8VMF三Y九W十VUTGN四二四66二五CKG十0L五1AaPGY五LATTYFbWPSHYBa94M8U44V一X十N三6Y90DeU08A七RCHNBRM23M8S0O2UROKScC7D五7",
	"九一cDN8S七UPNVP3七AON二IZP1T3九d1aQd二四I四b九L九六69LQ2十2TQN九Q一四d一QN七A二5A9b六AB3UBCG8Z四Ne四RX61bC五SVZY3一三GZL十七U一",
	"二2六六NK二cAJG59七BJK九DS7TOAe9C二九7U7U六M三1eRd八NEY6XJNFK七七OK66MBS五P6一P十9c5VRCce33IW7cAdV9SWHIWJ九十M3R十Lb12一",
	"W7PDBN6四4HUOVeLP二九ESS9CF88WA9MOZC0五VbJd5Bb三N2SY四KCQ七JS七三0d58A8九一7MLJJaKX2bcFQ5RKD9X7CY十2十6EU二d3Ld2VR",
	"84B四二HVRBGMQ6RBXKS六e八P九F2OHcS74DO三d八E1O八OJ十三一a一八1Q9YJ八四3RI十三三MdH四14KLa四82二五WU5a0B九QVW八Y五OM22二MVO二二a6",
	"dEbEBaGL五五9AU0K8DO9五XAMbD六Z六PE九十Q二M一C7J61F九94bMGX四MC四YD3ZTW6VWYO9A4七7UEaKe92PK9HdTM七D9PX61WQ九0AVUZ六2",
	"81MINIT十MOd二A三七KaDTe1b4S8RBXQNKGb00Nc四MdKMN54K二D2七DH8d6QC8五XJ9EA1CG七W7Q7SY9X三dC九F一bAN二QI六VObD五THP2bP",
	"L7N7Z三S4P一808DNLESEad8七PZZ324ab九5VC82QURU0I六JE七八6九8dIXAW4HQOO五MMVZY五KC0八Q3XWH1VFSeLT2YARPYLVReHAee7E",
	"QQT3五Z七GF71Td九L5S7dJc9b五Q十KD5五四XaM4J3二O3KW0MLAXcBIAJ2ReG5ED2aEYP6Y四7G7七三0YJ七aG七AV九JSbJ8LdUeAYP三b一七4J",
	"a6DICLU九2AHNO十四R七四Ue4五八Oa0E十7PU0CTTJ四CM79六YZZdDFN十K2G三D十三PCdSOQJ20bY七TRGFK04W五C6六9ER五7eZ七b八FAKNW二3C四",
	"2八5四KK四7MWOULWTZe九9BL一JFcF七Qb十Y五d九D1aQF2cTAA二54三LX6一Z二六四8JTZ七三八dcJ十八J1KeY三1一四CWCP7d四三7B七HV六四bC四N48J4",
	"四二5DbLBNaKRT9YJ六aX6EaE一6BN三四c6CPDF3T三WP三M1QdO五LQP八九L9O四7八1W06一XIHAUWGWYD三D9一CST0YFR十XE二二九b十六O93Z9M七一",
}

func TestHash(t *testing.T){
	var f = fnv.New64()
	f.Reset()
	f.Write([]byte(`1`))
	var res = f.Sum64()
	t.Log("结果是：", res)
}

func TestFilter_Push(t *testing.T) {
	var f = filter{
		Bytes: make([]byte, 1000),
		Hashes: []hash.Hash64{fnv.New64(), murmur3.New64()},
	}
	for _, v:= range testData{
		f.Push([]byte(v))
	}
	t.Log(f.Exists([]byte("234")))
	t.Log(f.Exists([]byte("I四八FHTQ2九四Zb7GT5McEIP1MWGLF五26SdGG7bSMFY五WSV8F5UVSC十C一5EQ六IULE十七IRO二四c4MVBP7VEHBJ8TX4一663HeJ四QZ九1L12")))
}

func TestGetFlasePositiveRate(t *testing.T) {
	t.Log(GetFlasePositiveRate(100, 100, 2))
	t.Log(GetFlasePositiveRate(1000, 100, 2))
}



func BenchmarkFilter_Push(b *testing.B) {
	var f = filter{
		Bytes: make([]byte, 1000),
		Hashes: []hash.Hash64{fnv.New64(), crc64.New( crc64.MakeTable(crc64.ISO))},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			f.Push([]byte("1234"))
		}
	})
	b.Log(runtime.NumGoroutine())
}

func TestWriteToRedis(t *testing.T){
	var Client = redis.NewClient(&redis.Options{
		Addr:     "192.168.30.156:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	var f = filter{
		Bytes: make([]byte, 1000),
		Hashes:  DefaultHash,
	}
	for _, v:= range testData{
		f.Push([]byte(v))
	}
	Client.Do("HSET", "test", "Bytes",f.Bytes, "AlreadyCount", 2 )
}

func TestNewRedisFilter(t *testing.T) {
	var rf, _ = NewRedisFilter("test", 1000, "192.168.30.156:6379","", 3, DefaultHash...)
	for _, v:= range testData{
		rf.Push([]byte(v))
	}
	rf.Write()
	t.Log(rf.Exists([]byte("234")))
	t.Log(rf.Exists([]byte("I四八FHTQ2九四Zb7GT5McEIP1MWGLF五26SdGG7bSMFY五WSV8F5UVSC十C一5EQ六IULE十七IRO二四c4MVBP7VEHBJ8TX4一663HeJ四QZ九1L12")))
}

func BenchmarkNewRedisFilter(b *testing.B) {
	var rf, err = NewRedisFilter("test", 1000, "192.168.30.156:6379","", 1, DefaultHash...)
	if err != nil {
		b.Fatal(err.Error())
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next(){
			rf.Push([]byte(`indexnext |previous |Redis 命令参考 » Hash（哈希表） »
HSET
HSET key field value

将哈希表 key 中的域 field 的值设为 value 。

如果 key 不存在，一个新的哈希表被创建并进行 HSET 操作。

如果域 field 已经存在于哈希表中，旧值将被覆盖。

可用版本：
>= 2.0.0
时间复杂度：
O(1)
返回值：
如果 field 是哈希表中的一个新建域，并且值设置成功，返回 1 。
如果哈希表中域 field 已经存在且旧值已被新值覆盖，返回 0 。
redis> HSET website google "www.g.cn"       # 设置一个新域
(integer) 1

redis> HSET website google "www.google.com" # 覆盖一个旧域
(integer) 0
indexnext |previous |Redis 命令参考 » Hash（哈希表） »
© Copyright 2013, Redis. Last updated on Dec 20, 2013. Created using Redis爱好者
  v: latest`))
		}
	})
	b.Log(runtime.NumGoroutine())
}
