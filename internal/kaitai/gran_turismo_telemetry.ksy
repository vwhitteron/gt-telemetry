meta:
  id: gran_turismo_telemetry
  title: Gran Turismo Telemetry
  license: CC0-1.0
  endian: le
  bit-endian: le
seq:
  - id: header
    type: header
    -doc: File header
  - id: map_position_coordinates
    type: coordinate
    -doc: Positional coordinates of vehicle on map in meters
  - id: velocity_vector
    type: vector
    -doc: Vehicle velocity vector in meters per second
  - id: rotation_axes
    type: symmetry_axes
    -doc: Body rotation axes (-1 to 1)
  - id: heading
    type: f4
    -doc: Orientation to North from 0.0(south) to 1.0(north).
  - id: angular_velocity_vector
    type: vector
    -doc: Angular velocity vector in radians per second (-1 to +1)
  - id: ride_height
    type: f4
    -doc: Vehicle ride height in meters
  - id: engine_rpm
    type: f4
    -doc: Engine speed in RPM
  - id: oiv
    type: f4
    -doc: Seed value for Salsa20 cipher, ignored
  - id: fuel_level
    type: f4
    -doc: Fuel remaining (0.0 to 1.0)
  - id: fuel_capacity
    type: f4
    -doc: Total fuel capacity (0.0 to 1.0)
  - id: ground_speed
    type: f4
    -doc: Vehicle ground speed in meters per second
  - id: manifold_pressure
    type: f4
    -doc: Manifold pressure in bar, only populated when turbo present (subtract 1 for boost pressure as negative values report vacuum)
  - id: oil_pressure
    type: f4
    -doc: Oil pressure
  - id: water_temperature
    type: f4
    -doc: Water temperature in celsius
  - id: oil_temperature
    type: f4
    -doc: Oil temperature in celsius
  - id: tyre_temperature
    type: corner_set
    -doc: Tyre temperatures in celsius
  - id: sequence_id
    type: u4
    -doc: Packet sequence ID
  - id: current_lap
    type: u2
    -doc: Current lap number
  - id: race_laps
    type: u2
    -doc: Total laps in race
  - id: best_laptime
    type: s4
    -doc: Personal best lap time for this session in milliseconds (-1ms when not set)
  - id: last_laptime
    type: s4
    -doc: Last lap time in milliseconds (-1ms when not set)
  - id: time_of_day
    type: u4
    -doc: Current time of day on track in milliseconds
  - id: starting_position
    type: s2
    -doc: Starting position at the beginning of the race (-1 when race starts)
  - id: race_entrants
    type: s2
    -doc: Total number of entrants at the beginning of the race (-1 when race starts)
  - id: rev_light_rpm_min
    type: u2
    -doc: Minimum engine RPM at which the shift light activates
  - id: rev_light_rpm_max
    type: u2
    -doc: Maximum engine RPM at which the shift light activates
  - id: calculated_max_speed
    type: u2
    -doc: Calculated maximum speed of the vehicle in kilometers per hour
  - id: flags
    type: flags
    -doc: Various flags for the current state of play and instrument cluster lights
  - id: transmission_gear
    type: transmission_gear
    -doc: Transmission gear selection
  - id: throttle
    type: u1
    -doc: Throttle position (0 to 255)
  - id: brake
    type: u1
    -doc: Brake position (0 to 255)
  - id: ignore_1
    size: 1
    -doc: Field 0x93 is empty and ignored
  - id: road_plane_vector
    type: vector
    -doc: Road plane vector
  - id: road_plane_distance
    type: u4
    -doc: Road plane distance
  - id: wheel_radians_per_second
    type: corner_set
    -doc: Individual wheel rotational speed in radians per second
  - id: tyre_radius
    type: corner_set
    -doc: Individual tyre radius in meters
  - id: suspension_height
    type: corner_set
    -doc: Individual suspension height at each corner in meters
  - id: reserved
    size: 32
    -doc: Reserved data, currently unused
  - id: clutch_actuation
    type: f4
    -doc: Clutch actuation (0.0 to 1.0)
  - id: clutch_engagement
    type: f4
    -doc: Clutch engagement (0.0 to 1.0)
  - id: cluch_output_rpm
    type: f4
    -doc: Rotational speed on the output side of the clutch in rpm
  - id: transmission_top_speed_ratio
    type: f4
    -doc: Ratio between vehicle top speed and wheel rotation speed (can calculate rpm at top speed and differential ratio)
  - id: transmission_gear_ratio
    type: gear_ratio
    -doc: Gear ratios for each gear in the transmission
  - id: vehicle_id
    type: u4
    -doc: ID of the vehicle
types:
  header:
    doc: Magic file header
    seq:
      - id: magic
        contents: [0x30, 0x53, 0x37, 0x47]
  vector:
    doc: Vector
    seq:
      - id: vector_x
        type: f4
      - id: vector_y
        type: f4
      - id: vector_z
        type: f4
  coordinate:
    doc: Vector
    seq:
      - id: coordinate_x
        type: f4
      - id: coordinate_y
        type: f4
      - id: coordinate_z
        type: f4
  symmetry_axes:
    doc: Symmetry axes
    seq:
      - id: pitch
        type: f4
      - id: yaw
        type: f4
      - id: roll
        type: f4
  corner_set:
    doc: Data set representing each wheel or suspension component at each corner of the vehicle
    seq:
      - id: front_left
        type: f4
      - id: front_right
        type: f4
      - id: rear_left
        type: f4
      - id: rear_right
        type: f4
  flags:
    doc: Various flags for the current state of play and instrument cluster lights
    seq:
      - id: live
        type: b1
      - id: game_paused
        type: b1
      - id: loading
        type: b1
      - id: in_gear
        type: b1
      - id: has_turbo
        type: b1
      - id: rev_limiter_alert
        type: b1
      - id: hand_brake_active
        type: b1
      - id: headlights_active
        type: b1
      - id: high_beam_active
        type: b1
      - id: low_beam_active
        type: b1
      - id: asm_active
        type: b1
      - id: tcs_active
        type: b1
      - id: flag13
        type: b1
      - id: flag14
        type: b1
      - id: flag15
        type: b1
      - id: flag16
        type: b1
  transmission_gear:
    doc: Transmission gear selection information
    seq:
      - id: current
        type: b4
      - id: suggested
        type: b4
  gear_ratio:
    doc: Gear ratios for each gear in the transmission
    seq:
      - id: gear
        type: f4
        repeat: expr
        repeat-expr: 8