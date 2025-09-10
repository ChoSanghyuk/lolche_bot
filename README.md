# Lolchebot



## 개요

롤체지지 홈페이지의 추천 덱을 하위 덱부터 순차적으로 깬다고 했을 때, 깬 덱을 기록하고 이번 차례에 할 덱을 텔레그램 봇을 통해 알려주는 코드입니다.



## 프로젝트 구조

```
  ├── bot.go                # lolchebot 구현
  ├── services.go           # lolchebot이 사용하는 interface
  ├── types.go              # 프로젝트 내 공통 변수 및 타입 정의
  ├── cmd/
  │   └── main.go           # Application 기동
  ├── config/
  │   ├── config.go         # 설정 파일 로직
  │   └── config.yaml       # 설정 파일
  ├── crawl/
  │   ├── crawler.go        # Web crawler 구현
  │   ├── crawler_test.go   # Crawler unit tests
  └── db/
      ├── db.go             # Db 접근 구현체
      ├── db_test.go        # Database unit tests
      └── model.go          # gorm struct
```



## 주요 동작



bot과의 상호작용은 Text Command와 Button Interaction을 통한 방법이 있다.

Text Commands:

  - /help → `helpJob()` - 모든 Text Commands 반환
  - /mode → `modeJob()` - 현재 모드 반환 (main 또는 pbe)
  - /switch → `switchJob()` - 모드 전환 (main <=> pre)
  - /update → `updateJob()` - 추천 덱을 크롤링 한 후, 완료한 덱을 필터링하여 현재 차례의 덱을 반환("일반 덱"/"증강 덱" interactive button 제공)
  - /reset → `resetJob()` - 완료 내역 전체 제거
  - /done → `doneJob()` - 완료 내역 반환 ("완료 목록" interactive button 제공)
  - /fix → `fixJob()` - 홈페이지의 css 코드가 변경되었을 때, 덱 크롤링 타겟 css path 자동 조정

  Button Interactions:
  - "일반 덱"/"증강 덱" buttons → `selectJob()` - 덱 상세 페이지 url과 "완료 여부" interactive button 제공
  - "완료 여부" button → `completeJob()` - 선택된 덱 완료 처리
  - "완료 목록" button → `restoreJob()` - 선택된 덱 완료 내역에서 제거
