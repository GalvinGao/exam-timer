# Exam Timer
> Timing the elapsed time you spent on specific questions in the exam, giving a better opportunity to win ;)

## Requirements
- Python 3
- Package `pynput`
- Package `strictyaml`

## How to use
1. Get [Python](https://python.org)
2. Install dependencies `pip install -r requirements.txt`
3. Edit `config.yml` with your needs
4. Run the program `python(3) timer.py`


## Keybinds
- `s`: Start session
- `(space)`: Next question
- `j`: Jump to question
- `p`: Pause/resume timer
- `e`: End session

## Directory Structure
**(The program will automatically create the required directories)**
```text
├─logs
│  └─ {date}_{exam_name}_section-{exam_section}.log
├─records
│  └─ {date}_{exam_name}_section-{exam_section}.json
└─reports
   └─ {date}_{exam_name}_section-{exam_section}.txt
```
- `logs` provides full record for the output of the program
- `records` provides JSON data (program-friendly) of the session
- `reports` provides a detailed description of the session

## TODO
- [] Make a real timer - the one will notify user when the time has run out
- [] GUI maybe?
- [] Rebuild (because currently the code just sucks)
